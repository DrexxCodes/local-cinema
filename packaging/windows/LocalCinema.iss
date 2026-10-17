; Inno Setup script for Local Cinema (needs Inno Setup 6.4 or newer).
; Built by .github/workflows/release.yml, for example:
;   ISCC.exe /DAppVersion=1.2.0 /DRepo=owner/local-cinema /DFfmpegBundle=2026.10 /DFfmpegSha256=<hash> LocalCinema.iss
; FFmpeg is not inside the installer. It is downloaded during setup.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef Repo
  #define Repo "OWNER/local-cinema"
#endif
#ifndef FfmpegBundle
  #define FfmpegBundle ""
#endif
#ifndef FfmpegSha256
  #define FfmpegSha256 ""
#endif
#ifndef SourceExe
  #define SourceExe "..\..\dist\local-cinema-gui.exe"
#endif

#define AppName "Local Cinema"
#define AppExe "local-cinema.exe"
#define Port "8765"
#define DataRoot "{commonappdata}\Local Cinema"

[Setup]
AppId={{B7E4C1A2-5D3F-4E8A-9C61-2F0A7D8E4B13}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher=Drexx Technologies
AppPublisherURL=https://github.com/{#Repo}
AppSupportURL=https://github.com/{#Repo}/issues
AppUpdatesURL=https://github.com/{#Repo}/releases
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
PrivilegesRequired=admin
MinVersion=10.0
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\..\dist\installer
OutputBaseFilename=LocalCinema-Setup-{#AppVersion}
SetupIconFile=..\icon\icon.ico
UninstallDisplayIcon={app}\{#AppExe}
LicenseFile=..\..\LICENSE.md
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
CloseApplications=yes
RestartApplications=no

[Tasks]
Name: "startup"; Description: "Start Local Cinema when Windows starts"
Name: "desktopicon"; Description: "Create a desktop shortcut"; Flags: unchecked

[Dirs]
; Data lives outside the program folder, so updates and uninstalls leave it alone.
Name: "{#DataRoot}"; Permissions: users-modify
Name: "{#DataRoot}\data"; Permissions: users-modify
Name: "{#DataRoot}\bin"; Permissions: users-modify
Name: "{#DataRoot}\logs"; Permissions: users-modify

[Files]
Source: "{#SourceExe}"; DestDir: "{app}"; DestName: "{#AppExe}"; Flags: ignoreversion
Source: "..\..\LICENSE.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\THIRD_PARTY_NOTICES.md"; DestDir: "{app}"; Flags: ignoreversion

[InstallDelete]
; Older versions made a browser shortcut file here.
Type: files; Name: "{app}\Open Local Cinema.url"

[Icons]
; The program opens its control panel. Starting it again just brings that window to the front.
Name: "{group}\Open Local Cinema"; Filename: "{app}\{#AppExe}"
Name: "{group}\Stop Local Cinema"; Filename: "{sys}\taskkill.exe"; Parameters: "/im {#AppExe} /f"; IconFilename: "{app}\{#AppExe}"; Flags: runminimized
Name: "{group}\Uninstall Local Cinema"; Filename: "{uninstallexe}"
Name: "{autodesktop}\Local Cinema"; Filename: "{app}\{#AppExe}"; Tasks: desktopicon
; At log in it starts minimized, so the cinema is running without a window in the way.
Name: "{commonstartup}\Local Cinema"; Filename: "{app}\{#AppExe}"; Parameters: "--background"; Tasks: startup

[Run]
; Open the firewall so phones and TVs on the network can reach the cinema.
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Local Cinema"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Local Cinema"" dir=in action=allow protocol=TCP localport={#Port} profile=any"; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Local Cinema discovery"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Local Cinema discovery"" dir=in action=allow protocol=UDP localport=5353 profile=any"; Flags: runhidden
; Start the app. It opens the control panel, and on a fresh install it opens the setup page in the browser.
Filename: "{app}\{#AppExe}"; Description: "Start Local Cinema now"; Flags: nowait postinstall skipifsilent runasoriginaluser

[UninstallRun]
Filename: "{sys}\taskkill.exe"; Parameters: "/im {#AppExe} /f"; Flags: runhidden; RunOnceId: "StopApp"
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Local Cinema"""; Flags: runhidden; RunOnceId: "FirewallRule"
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Local Cinema discovery"""; Flags: runhidden; RunOnceId: "FirewallRule2"

[UninstallDelete]
Type: files; Name: "{app}\installed"

[Code]
var
  DownloadPage: TDownloadWizardPage;

function FfmpegDir: String;
begin
  Result := ExpandConstant('{commonappdata}\Local Cinema\bin');
end;

function FfmpegPresent: Boolean;
begin
  Result := FileExists(FfmpegDir + '\ffmpeg.exe') and FileExists(FfmpegDir + '\ffprobe.exe');
end;

function OnDownloadProgress(const Url, FileName: String; const Progress, ProgressMax: Int64): Boolean;
begin
  if Progress = ProgressMax then
    Log(Format('Downloaded %s', [FileName]));
  Result := True;
end;

procedure InitializeWizard;
begin
  DownloadPage := CreateDownloadPage('Installing necessary components',
    'We are installing some necessary components for Local Cinema', @OnDownloadProgress);
end;

// Unpacks the FFmpeg zip with the PowerShell that ships with Windows 10 and 11.
function UnpackFfmpeg(const ZipFile: String): Boolean;
var
  ResultCode: Integer;
  Zip, Dest: String;
begin
  Zip := ZipFile;
  Dest := FfmpegDir;
  StringChangeEx(Zip, '''', '''''', True);
  StringChangeEx(Dest, '''', '''''', True);
  ForceDirectories(FfmpegDir);
  Result := Exec(ExpandConstant('{sys}\WindowsPowerShell\v1.0\powershell.exe'),
    '-NoProfile -ExecutionPolicy Bypass -Command "Expand-Archive -LiteralPath ''' + Zip + ''' -DestinationPath ''' + Dest + ''' -Force"',
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
end;

function NextButtonClick(CurPageID: Integer): Boolean;
var
  Again: Boolean;
begin
  Result := True;
  if (CurPageID = wpReady) and (not FfmpegPresent) and ('{#FfmpegBundle}' <> '') then
  begin
    repeat
      Again := False;
      DownloadPage.Clear;
      DownloadPage.Add('https://github.com/{#Repo}/releases/download/ffmpeg-{#FfmpegBundle}/ffmpeg-windows-amd64.zip',
        'ffmpeg.zip', '{#FfmpegSha256}');
      DownloadPage.SetText('Downloading video tools (FFmpeg)', 'This can take a few minutes on a slow connection.');
      DownloadPage.Show;
      try
        try
          DownloadPage.Download;
          DownloadPage.SetText('Unpacking video tools', '');
          if not UnpackFfmpeg(ExpandConstant('{tmp}\ffmpeg.zip')) then
            RaiseException('The video tools could not be unpacked.');
        except
          if DownloadPage.AbortedByUser then
            Result := False
          else
            case SuppressibleMsgBox(AddPeriod(GetExceptionMessage) + #13#10#13#10 +
              'Local Cinema needs FFmpeg for thumbnails and for converting movies. ' +
              'Choose Retry to try again, or Ignore to continue without it (you can install it later in Settings).',
              mbError, MB_ABORTRETRYIGNORE, IDIGNORE) of
              IDRETRY: Again := True;
              IDABORT: Result := False;
            end;
        end;
      finally
        DownloadPage.Hide;
      end;
    until not Again;
  end;
end;

// Stops a running copy before files are replaced during an update.
function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/im {#AppExe} /f', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Result := '';
end;

// The marker tells the program it is installed, so it keeps its data in ProgramData.
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
    SaveStringToFile(ExpandConstant('{app}\installed'), 'installed by the Local Cinema setup', False);
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    if (not UninstallSilent) and (MsgBox('Do you also want to delete your Local Cinema data (rooms, settings, thumbnails and converted movies)?' + #13#10#13#10 +
      'Choose No to keep it for a future install.', mbConfirmation, MB_YESNO or MB_DEFBUTTON2) = IDYES) then
      DelTree(ExpandConstant('{commonappdata}\Local Cinema'), True, True, True)
    else
      DelTree(FfmpegDir, True, True, True); // the video tools can be downloaded again
  end;
end;
