export type TourStep = {
  id: string;
  /** Matches a data-tour attribute in the player. No target means a plain message. */
  target?: string;
  title: string;
  text: string;
  /** Only shown on touch screens. */
  touchOnly?: boolean;
};

export const FIRST_TIME_KEY = "isFirstTime";
export const IOS_HINT_KEY = "iosFullscreenHintDismissed";

export const TOUR_STEPS: TourStep[] = [
  { id: "welcome", title: "Welcome to the player", text: "Here is a quick tour of the controls. Use the arrows to move through it, or skip it at any time." },
  { id: "play", target: "play", title: "Play and pause", text: "Starts or pauses the movie. On a computer you can also click the picture or press the space bar." },
  { id: "back", target: "back", title: "Back 10 seconds", text: "Missed a line? Jump back ten seconds." },
  { id: "forward", target: "forward", title: "Forward 10 seconds", text: "Skip ahead ten seconds." },
  { id: "volume", target: "volume", title: "Volume", text: "Mute or change the volume. On a phone, use the volume buttons on the side." },
  { id: "seek", target: "seek", title: "Seek bar", text: "Drag to jump to any moment. The brighter part shows how much has loaded, so you can skip ahead smoothly." },
  { id: "speed", target: "speed", title: "Playback speed", text: "Watch slower or faster, from half speed to double." },
  { id: "captions", target: "captions", title: "Subtitles", text: "Turn subtitles on or off and pick a language, when the movie has any." },
  { id: "rotate", target: "rotate", title: "Rotate the screen", text: "Turns the picture sideways so it fills a phone held upright. Tap again for the other direction, and once more to go back." },
  { id: "fullscreen", target: "fullscreen", title: "Full screen", text: "Fills the whole screen." },
  { id: "more", target: "more", title: "More options", text: "Copy a stream URL to watch in another app such as VLC, plus picture in picture and AirPlay when your device supports them." },
  { id: "gestures", title: "Touch shortcuts", text: "Double-tap the left or right side of the picture to skip ten seconds. Tap once to show or hide the controls.", touchOnly: true },
  { id: "done", title: "You are ready", text: "Enjoy the movie! The controls are always at the bottom of the player." },
];

export const IOS_FULLSCREEN_NOTE = " On iPhone, full screen uses Apple's own controls. To watch with your phone sideways, turn off Portrait Orientation Lock in Control Center.";
