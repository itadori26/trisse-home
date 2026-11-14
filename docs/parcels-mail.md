# Parcels from shipping e-mails

The parcels widget can find tracking numbers in your shipping e-mails by itself (#166). trisse-home reads **one folder** of your mailbox over IMAP and adds every new parcel it finds.

## What it reads, and what it never does

- **Read-only.** The folder is opened with `EXAMINE`, so the server refuses any change, and messages are fetched with `BODY.PEEK`. Nothing is moved, deleted or flagged, not even marked as read.
- **Only new messages.** The first run looks back 14 days; after that only messages that arrived since the last run are read, at most 50 per run.
- **Only what it needs.** It reads the sender, the subject and the first 256 KB of the text (HTML reduced to its text and links). Attachments are skipped.
- **Nothing is stored except the result.** Only the tracking number, the carrier and a label (shop name or subject) are saved, never the e-mail.
- **Only IMAP over TLS** (implicit TLS, usually port 993). The app password is stored encrypted with `SECRET_KEY` and is never returned by the API.

## Setup

1. **Create an app password.** Most providers offer one for mail apps, e.g. Gmail, GMX, web.de, Outlook, Posteo or mailbox.org. Don't use your main password.
2. **Create a folder**, e.g. `Pakete`, and a **filter rule** that moves shipping mails into it. A rule with these conditions catches most of them:
   - sender contains `dhl.de`, `hermesworld`, `myhermes`, `dpd.de`, `gls-pakete`, `gls-group` or `ups.com`, **or**
   - subject contains `Sendung`, `Versand`, `versandt`, `shipped`, `on its way` or `unterwegs`.
3. **Optional: turn on DHL parcel announcements.** In your DHL customer account under _Einstellungen → Benachrichtigungen_, enable e-mail notifications. DHL then mails you about every parcel addressed to you, even when the shop doesn't send a tracking number.
4. In trisse-home open **Settings → Parcels → Parcels from e-mails**:
   - enter the server (e.g. `imap.gmx.net`), user name, app password and folder;
   - choose how often to check (every 15 minutes by default);
   - switch it on and save. The connection is tested before saving.

## Carriers

| Carrier | Recognised                                                                            | Tracking in the widget            |
| ------- | ------------------------------------------------------------------------------------- | --------------------------------- |
| DHL     | Sendungsnummer, JJD numbers, dhl.de links (check digit verified for 20-digit numbers) | Live status with your DHL API key |
| Hermes  | Sendungsnummer in Hermes mails, myhermes/hermesworld links                            | Link to the Hermes tracking page  |
| DPD     | Paketnummer in DPD mails, tracking.dpd.de links                                       | Link to the DPD tracking page     |
| GLS     | Paketnummer in GLS mails, gls-pakete.de links                                         | Link to the GLS tracking page     |
| UPS     | `1Z…` numbers anywhere (check digit verified)                                         | Link to the UPS tracking page     |

A number only counts when it appears right after a word like _Sendungsnummer_, _Paketnummer_ or _tracking number_, or inside a carrier's tracking link. Order numbers, phone numbers and IBANs don't match. When a shop mail doesn't say which carrier ships, the number is skipped instead of guessed.

## Removing a parcel

A parcel found in a mail shows an envelope ✉. If you remove it, it is remembered and not added again by a later run.

## Troubleshooting

The settings show when the mailbox was last read and the last error, e.g.:

- `login failed`: wrong user name or app password;
- `folder "Pakete"`: the folder doesn't exist (names are case-sensitive, sub-folders are often written `INBOX/Pakete` or `INBOX.Pakete`);
- `TLS`: the server or port doesn't offer IMAP over TLS (STARTTLS on port 143 is not supported).
