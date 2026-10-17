package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type serverState int

const (
	stateStopped serverState = iota
	stateStarting
	stateRunning
)

// server runs the cinema as a hidden child process: this same program started with
// `serve`. A Windows job object makes sure the child never outlives the panel, even if the
// panel crashes.
type server struct {
	exe, port string
	job       windows.Handle
	logf      func(string)
	onState   func(state serverState, detail string)

	mu            sync.Mutex
	cmd           *exec.Cmd
	state         serverState
	stoppedByUser bool
}

func newServer(port string, logf func(string), onState func(serverState, string)) (*server, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot find the program file: %w", err)
	}
	s := &server{exe: exe, port: port, logf: logf, onState: onState}
	job, err := newKillOnCloseJob()
	if err != nil {
		logf("note: could not create a job object (" + err.Error() + "), the server may keep running if the panel crashes")
	} else {
		s.job = job
	}
	return s, nil
}

func newKillOnCloseJob() (windows.Handle, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, err
	}
	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info)))
	if err != nil {
		_ = windows.CloseHandle(job)
		return 0, err
	}
	return job, nil
}

// State returns the current state.
func (s *server) State() serverState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Start launches the server. It does nothing when the server is already running.
func (s *server) Start() {
	s.mu.Lock()
	if s.state != stateStopped {
		s.mu.Unlock()
		return
	}
	s.state = stateStarting
	s.stoppedByUser = false
	s.mu.Unlock()
	s.onState(stateStarting, "")
	s.logf("Starting the server on port " + s.port + "...")

	cmd := exec.Command(s.exe, "serve", "--http=0.0.0.0:"+s.port)
	cmd.Env = append(os.Environ(), "CINEMA_CHILD=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	r, w, err := os.Pipe()
	if err != nil {
		s.failed("could not create a pipe: " + err.Error())
		return
	}
	cmd.Stdout, cmd.Stderr = w, w
	if err := cmd.Start(); err != nil {
		_ = r.Close()
		_ = w.Close()
		s.failed("could not start the server: " + err.Error())
		return
	}
	_ = w.Close() // the child keeps its own copy
	s.assignToJob(cmd)

	s.mu.Lock()
	s.cmd = cmd
	s.mu.Unlock()

	go func() {
		readLines(r, s.logf)
		_ = r.Close()
	}()
	go func() {
		err := cmd.Wait()
		s.mu.Lock()
		byUser := s.stoppedByUser
		s.cmd = nil
		s.state = stateStopped
		s.mu.Unlock()

		detail := "stopped"
		switch {
		case byUser:
			s.logf("The server was stopped.")
		case err != nil:
			detail = "it stopped unexpectedly: " + err.Error()
			s.logf("The server stopped unexpectedly: " + err.Error())
		default:
			s.logf("The server exited.")
		}
		s.onState(stateStopped, detail)
	}()
	go s.waitUntilReady(cmd)
}

func (s *server) failed(reason string) {
	s.mu.Lock()
	s.state = stateStopped
	s.mu.Unlock()
	s.logf("Error: " + reason)
	s.onState(stateStopped, reason)
}

func (s *server) assignToJob(cmd *exec.Cmd) {
	if s.job == 0 {
		return
	}
	h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return
	}
	defer windows.CloseHandle(h)
	if err := windows.AssignProcessToJobObject(s.job, h); err != nil {
		s.logf("note: could not add the server to the job object: " + err.Error())
	}
}

// waitUntilReady marks the server as running once it answers.
func (s *server) waitUntilReady(cmd *exec.Cmd) {
	for i := 0; i < 240; i++ { // about a minute
		s.mu.Lock()
		same := s.cmd == cmd && s.state != stateStopped
		s.mu.Unlock()
		if !same {
			return
		}
		if Ping(s.port) {
			s.mu.Lock()
			ok := s.cmd == cmd && s.state == stateStarting
			if ok {
				s.state = stateRunning
			}
			s.mu.Unlock()
			if ok {
				s.logf("The server is running at " + BaseURL(s.port))
				s.onState(stateRunning, "")
			}
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	s.logf("The server did not answer after a minute. Check the messages above.")
}

// Stop ends the server and everything it started (such as FFmpeg).
func (s *server) Stop() {
	s.mu.Lock()
	cmd := s.cmd
	s.stoppedByUser = true
	s.mu.Unlock()
	if cmd == nil {
		return
	}
	s.logf("Stopping the server...")
	if s.job == 0 || windows.TerminateJobObject(s.job, 0) != nil {
		_ = cmd.Process.Kill()
	}
}
