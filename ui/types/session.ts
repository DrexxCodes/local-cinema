export type Session = {
  token: string;
  session_id: string;
  user: { id: string; room_number: string; display_name: string };
};

export type CinemaInfo = { name: string; tagline: string; accent_color: string; setup_complete: boolean };
