# Security policy

## Supported versions

Security fixes are made for the latest release only. Please update before reporting
a problem, using Settings, Check for updates, or the newest installer on the
[releases page](../../releases).

## Reporting a vulnerability

Please report security problems privately, not in a public issue.

- Use GitHub's **Report a vulnerability** button on the repository's Security tab, or
- email **hell@drexxtech.ng** with the subject "Local Cinema security".

Include what you found, the version (Settings, About and updates), steps to repeat it,
and what an attacker could do with it. You can expect a first reply within a few
days. We will work on a fix, tell you when it is released, and credit you in the
release notes if you want that.

## What Local Cinema is designed for

Local Cinema is built for a **trusted local network**: a hotel, a school, a home or an
office, where the people connecting are guests you have given a profile login. It is not
hardened for the open internet. Please keep these points in mind:

- **Do not expose it to the internet.** Do not forward its port on your router. If you
  need remote access, use a VPN you control.
- **The admin panel is powerful.** Use a strong admin password and do not share it.
  Admin access can read your movie folders' file names and change who can watch.
- **Profile passwords are shared secrets.** Profiles are limited by the number of devices you
  allow. Change a password in the Users tab if it leaks.
- **Stream URLs act like passwords while they work.** A link made with "Copy stream URL"
  carries a key, and anyone who has it can watch that movie until the session ends
  (about 30 minutes after the player stops). It also counts against the profile's device
  limit. Turn external streaming off per profile if you do not want guests to use it.
- **The connection is plain HTTP by default.** Anyone who can watch traffic on your
  network can read it. This is normal for a closed local network.
- **Movie folders are only read.** Local Cinema never writes into your movie folders.
  Thumbnails, converted copies and subtitles you add are stored in its own data folder.
- **Downloads are verified.** The FFmpeg download is checked against a SHA-256 checksum
  before it is used. Updates are downloaded by you from the GitHub releases page, with a
  `checksums.txt` for every file.

## Out of scope

Reports about running the app on a public network, missing HTTPS on a local network,
or weak passwords chosen by the admin are not vulnerabilities, but we are happy to
improve the guidance above.
