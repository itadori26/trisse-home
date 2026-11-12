import type { ServerMessageKey } from '../../api-client/generated/server-messages';

/**
 * German translations of the backend error messages (#102). Keys come from
 * the generated catalogue; the type makes a missing translation a compile
 * error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A bookmark can have at most 10 tags.
  srv_350d099c21ede260: 'Ein Lesezeichen kann höchstens 10 Tags haben.',
  // A guest with this name already exists.
  srv_083c4fe6f2b25f45: 'Es gibt bereits einen Gast mit diesem Namen.',
  // A maintenance window must end after it starts and last at most 7 days.
  srv_a912b3e2ea2b8170: 'Ein Wartungsfenster muss nach seinem Beginn enden und darf höchstens 7 Tage dauern.',
  // A note needs a title or some text.
  srv_cf47637ff80f03dc: 'Eine Notiz braucht einen Titel oder etwas Text.',
  // A password or access token is required for new registries.
  srv_865d0fdde8c133ca: 'Für neue Registries ist ein Passwort oder Zugriffstoken erforderlich.',
  // A shortcut with this keyword already exists.
  srv_76a5625367cbe997: 'Es gibt bereits ein Kürzel mit diesem Schlüsselwort.',
  // Actions are not enabled for this container (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).
  srv_fc1c5d36ee0973d3:
    'Aktionen sind für diesen Container nicht freigeschaltet (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).',
  // AdGuard Home is not configured yet.
  srv_9bc99814ae6ec519: 'AdGuard Home ist noch nicht eingerichtet.',
  // AdGuard Home is not reachable.
  srv_66e0c4283a72cd94: 'AdGuard Home ist nicht erreichbar.',
  // An API token with the results:read scope is required.
  srv_d05b6c7e6905729f: 'Ein API-Token mit dem Bereich results:read ist erforderlich.',
  // At most 10 calendars are supported.
  srv_326bb72e415244d3: 'Es werden höchstens 10 Kalender unterstützt.',
  // At most 10 rules per tab are supported.
  srv_9c62e19aa7e4b027: 'Pro Tab werden höchstens 10 Regeln unterstützt.',
  // At most 20 registries can be stored.
  srv_2ad8dffbd734017a: 'Es können höchstens 20 Registries gespeichert werden.',
  // At most 20 rules are supported.
  srv_741879779a784a6e: 'Es werden höchstens 20 Regeln unterstützt.',
  // At most 30 parcels can be tracked.
  srv_5343c52cdf9d2afd: 'Es können höchstens 30 Sendungen verfolgt werden.',
  // At most 50 templates can be stored.
  srv_fb588d9d1c9ff8e5: 'Es können höchstens 50 Vorlagen gespeichert werden.',
  // At most {max} bookmarks can be imported at once.
  srv_c40eff143d7c40cd: 'Es können höchstens {max} Lesezeichen auf einmal importiert werden.',
  // Between 1 and 365 backups can be kept.
  srv_4bba6d234928bd07: 'Es können zwischen 1 und 365 Backups aufbewahrt werden.',
  // Choose a backup file.
  srv_56aef011f47cdbeb: 'Bitte eine Backup-Datei auswählen.',
  // Choose at least one scope.
  srv_c2c503807749b722: 'Bitte mindestens einen Bereich wählen.',
  // Choose between 0 (never) and 3650 days.
  srv_0054eff4d1acb233: 'Bitte zwischen 0 (nie) und 3650 Tagen wählen.',
  // Choose what the maintenance applies to.
  srv_4e9dc8e29fa35fbf: 'Bitte wählen, wofür die Wartung gilt.',
  // Colors must look like #1a2b3c.
  srv_100561a937e88349: 'Farben müssen wie #1a2b3c aussehen.',
  // Could not connect to AdGuard Home: {detail}
  srv_a3e65fea5b434e1e: 'Keine Verbindung zu AdGuard Home: {detail}',
  // Could not connect to Nextcloud: {detail}
  srv_6d1befe42898ea11: 'Keine Verbindung zu Nextcloud: {detail}',
  // Could not connect to Speedtest Tracker: {detail}
  srv_0308331877b78de4: 'Keine Verbindung zu Speedtest Tracker: {detail}',
  // Could not connect to the FRITZ!Box: {detail}
  srv_6a8ef763439ca559: 'Keine Verbindung zur FRITZ!Box: {detail}',
  // Could not load the calendar: {detail}
  srv_1eb130fb601040a1: 'Der Kalender konnte nicht geladen werden: {detail}',
  // Could not load the waste calendar: {detail}
  srv_57ceb20ccebc67bb: 'Der Abfallkalender konnte nicht geladen werden: {detail}',
  // Could not read the task list: {detail}
  srv_b8061d9b619a4928: 'Die Aufgabenliste konnte nicht gelesen werden: {detail}',
  // DHL does not know this tracking number (yet).
  srv_24f489be60f9182b: 'DHL kennt diese Sendungsnummer (noch) nicht.',
  // DHL rejected the API key.
  srv_099fe605093d8b87: 'DHL hat den API-Schlüssel abgelehnt.',
  // Delivered parcels can be kept for 0 to 30 days.
  srv_7cb71b2e113119ec: 'Zugestellte Sendungen können 0 bis 30 Tage angezeigt werden.',
  // Each registry may only be listed once.
  srv_4ae76664f7b46c6b: 'Jede Registry darf nur einmal eingetragen sein.',
  // Each registry needs a host name (e.g. ghcr.io) and a user name.
  srv_e9bcefee5bd2a359: 'Jede Registry braucht einen Hostnamen (z. B. ghcr.io) und einen Benutzernamen.',
  // Enter your authentication code to finish signing in.
  srv_1eb555800ce2927d: 'Gib deinen Authentifizierungscode ein, um die Anmeldung abzuschließen.',
  // Events can be at most one year long.
  srv_38ba62864162fcec: 'Termine können höchstens ein Jahr lang sein.',
  // Every calendar URL must start with http:// or https://.
  srv_ffb97fcecc45c54a: 'Jede Kalender-URL muss mit http:// oder https:// beginnen.',
  // Every rule needs a match text of at most 40 characters.
  srv_8e50b118b5e6b205: 'Jede Regel braucht einen Suchtext mit höchstens 40 Zeichen.',
  // Export version {version} is not supported by this version of trisse-home.
  srv_43fe335aacf98b3c: 'Exportversion {version} wird von dieser trisse-home-Version nicht unterstützt.',
  // Guests can only view this dashboard.
  srv_79bf20d17f6aa6eb: 'Gäste können dieses Dashboard nur ansehen.',
  // Host metrics are not available. Are /proc and /sys mounted?
  srv_b854584a505d6c89: 'Host-Messwerte sind nicht verfügbar. Sind /proc und /sys eingebunden?',
  // Icons must be at most 8 characters long.
  srv_72f2ef28f00d9995: 'Icons dürfen höchstens 8 Zeichen lang sein.',
  // Invalid scope {scope}: use module:read or module:write.
  srv_8648b1ab32906623: 'Ungültiger Bereich {scope}: bitte modul:read oder modul:write verwenden.',
  // Invalid scope {scope}: use service:<name>, container:<name> or *.
  srv_d6e30a440a4e7b86: 'Ungültige Angabe {scope}: bitte service:<Name>, container:<Name> oder * verwenden.',
  // Items must be between 1 and 200 characters long.
  srv_593a1c118a3d9d67: 'Einträge müssen zwischen 1 und 200 Zeichen lang sein.',
  // Labels must be at most 40 characters long.
  srv_4cd4b235493d7de0: 'Bezeichnungen dürfen höchstens 40 Zeichen lang sein.',
  // Latitude and longitude are required.
  srv_9d350adc128ef832: 'Breiten- und Längengrad sind erforderlich.',
  // Login failed: {detail}
  srv_53c1eb17a03496b2: 'Anmeldung fehlgeschlagen: {detail}',
  // Missing or invalid CSRF token.
  srv_eccd19aa12e79824: 'CSRF-Token fehlt oder ist ungültig.',
  // Nextcloud is not configured yet.
  srv_852391e96f8a11b2: 'Nextcloud ist noch nicht eingerichtet.',
  // No MAC address is stored for this service.
  srv_591001f7788fb502: 'Für diesen Dienst ist keine MAC-Adresse hinterlegt.',
  // No S.M.A.R.T. export found. Run scripts/smart-export.sh on the host (see README).
  srv_74822a0f2779c188:
    'Kein S.M.A.R.T.-Export gefunden. Bitte scripts/smart-export.sh auf dem Host ausführen (siehe README).',
  // No WireGuard export found. Run scripts/wireguard-export.sh on the host (see README).
  srv_9d762d7de49b8d86:
    'Kein WireGuard-Export gefunden. Führe scripts/wireguard-export.sh auf dem Host aus (siehe README).',
  // No account has been set up yet.
  srv_a53afcf337b06aea: 'Es wurde noch kein Konto eingerichtet.',
  // No bookmarks found. Please export your bookmarks as HTML from the browser.
  srv_2e56580cd044b5c5: 'Keine Lesezeichen gefunden. Bitte exportiere deine Lesezeichen im Browser als HTML.',
  // No calendar is configured yet.
  srv_92fe9be68e4fbeb4: 'Es ist noch kein Kalender eingerichtet.',
  // No off-site backup target is configured.
  srv_7f26dde75768b4ec: 'Es ist kein externes Backup-Ziel eingerichtet.',
  // No waste calendar is configured yet.
  srv_d1f1feb231209e5a: 'Es ist noch kein Abfallkalender eingerichtet.',
  // Only the owner can do this.
  srv_67f0ba90e3a4e5b4: 'Das kann nur der Besitzer.',
  // Only the owner can restore backups in the browser.
  srv_2fa2f3b7606fee5e: 'Nur der Besitzer kann Backups im Browser wiederherstellen.',
  // Passkeys need HTTPS and PUBLIC_URL (see the README).
  srv_b90b21f1d731b4f6: 'Passkeys brauchen HTTPS und PUBLIC_URL (siehe README).',
  // Password login is disabled. Use single sign-on.
  srv_fd5d4b3fb31aa85f: 'Die Anmeldung mit Passwort ist deaktiviert. Bitte Single Sign-on verwenden.',
  // Please add your DHL API key in the settings first.
  srv_7d0f681f466f8e83: 'Bitte zuerst deinen DHL-API-Schlüssel in den Einstellungen hinterlegen.',
  // Please choose a stop in the widget settings.
  srv_d92a0962a35bc0f4: 'Bitte wähle in den Widget-Einstellungen eine Haltestelle.',
  // Please choose one of the configured calendars.
  srv_f5627673a11b56bd: 'Bitte wähle einen der eingerichteten Kalender.',
  // Please enter a valid sender address.
  srv_ac8a8ae2b384e281: 'Bitte eine gültige Absenderadresse eingeben.',
  // Please enter a valid start and end.
  srv_b71f32ebd27a0b6a: 'Bitte gib einen gültigen Beginn und ein gültiges Ende ein.',
  // Please enter and save a topic first.
  srv_dd3a700b0be13baf: 'Bitte zuerst ein Topic eingeben und speichern.',
  // Please enter at least one calendar URL.
  srv_a4d2757d3ca7d004: 'Bitte mindestens eine Kalender-URL eingeben.',
  // Please enter between 2 and 100 characters.
  srv_990b607e23145ec7: 'Bitte zwischen 2 und 100 Zeichen eingeben.',
  // Please enter the SMTP server name.
  srv_5057fd283cbaf8f6: 'Bitte den Namen des SMTP-Servers eingeben.',
  // Please enter the code from your authenticator app or a recovery code.
  srv_2c825edcb1338a1c:
    'Bitte gib den Code aus deiner Authenticator-App oder einen Wiederherstellungscode ein.',
  // Please enter the password of the FRITZ!Box user.
  srv_eabd506dfb319855: 'Bitte gib das Passwort des FRITZ!Box-Benutzers ein.',
  // Please enter valid recipient addresses, separated by commas.
  srv_e9295c1002fdd2ba: 'Bitte gültige Empfängeradressen eingeben, durch Kommas getrennt.',
  // Quiet hours need a start and an end such as 22:00 and 07:00.
  srv_04eab3a4b4f8b837: 'Ruhezeiten brauchen einen Beginn und ein Ende, z. B. 22:00 und 07:00.',
  // Rules need weekdays 0–6, times like 07:30 and networks like 192.168.178.0/24.
  srv_560f31f8940f8cfc: 'Regeln brauchen Wochentage 0–6, Uhrzeiten wie 07:30 und Netze wie 192.168.178.0/24.',
  // Session not found.
  srv_785eadcc6dca712c: 'Sitzung nicht gefunden.',
  // Single sign-on is not configured.
  srv_c1cbece9dc9bd483: 'Single Sign-on ist nicht eingerichtet.',
  // Speedtest Tracker is not configured yet.
  srv_6f32001a15a15c4d: 'Speedtest Tracker ist noch nicht eingerichtet.',
  // Speedtest Tracker is not reachable.
  srv_cfc88a9871753350: 'Speedtest Tracker ist nicht erreichbar.',
  // Start the two-factor setup first.
  srv_32f91240e6490e41: 'Bitte zuerst die Einrichtung der Zwei-Faktor-Authentifizierung starten.',
  // Storing credentials is disabled in the demo.
  srv_ea4ce39d77587136: 'Das Speichern von Zugangsdaten ist in der Demo deaktiviert.',
  // TLS must be starttls, tls or none.
  srv_8ba7cc04f14ed353: 'TLS muss starttls, tls oder none sein.',
  // Tags may only contain letters, digits, - and _ (max. 30 characters).
  srv_7cdf99ade35d99e3: 'Tags dürfen nur Buchstaben, Ziffern, - und _ enthalten (max. 30 Zeichen).',
  // The Docker API is not reachable. Does the socket proxy allow IMAGES=1?
  srv_1120b37481579f63: 'Die Docker-API ist nicht erreichbar. Erlaubt der Socket-Proxy IMAGES=1?',
  // The FRITZ!Box is not configured yet.
  srv_6f0fbfbef4b05402: 'Die FRITZ!Box ist noch nicht eingerichtet.',
  // The ICS URL must start with https://, http:// or webcal://.
  srv_457b52382d260269: 'Die ICS-URL muss mit https://, http:// oder webcal:// beginnen.',
  // The URL must contain {query}.
  srv_1605df454fade766: 'Die URL muss {query} enthalten.',
  // The URL must look like http://192.168.1.2:3000.
  srv_ff20bffaaad40a7b: 'Die URL muss wie http://192.168.1.2:3000 aussehen.',
  // The URL must look like http://192.168.1.2:8765.
  srv_a74063c5a9d4db02: 'Die URL muss wie http://192.168.1.2:8765 aussehen.',
  // The URL must look like http://192.168.178.1:49000.
  srv_d14b64d005d328ef: 'Die URL muss wie http://192.168.178.1:49000 aussehen.',
  // The URL must look like https://cloud.example.com.
  srv_c32e8ec6145cd087: 'Die URL muss wie https://cloud.example.com aussehen.',
  // The URL must start with http:// or https://.
  srv_33e6dad254bc6264: 'Die URL muss mit http:// oder https:// beginnen.',
  // The Wake-on-LAN relay did not answer. Is COMPOSE_PROFILES=wol set (see README)?
  srv_7bd9cc89bdd1560a:
    'Das Wake-on-LAN-Relay hat nicht geantwortet. Ist COMPOSE_PROFILES=wol gesetzt (siehe README)?',
  // The WebDAV URL must look like https://cloud.example.org/remote.php/dav/files/USER/Backups/.
  srv_aa51a8ab465628a7:
    'Die WebDAV-URL muss wie https://cloud.example.org/remote.php/dav/files/BENUTZER/Backups/ aussehen.',
  // The WebDAV target is not usable: {detail}
  srv_91c4369358de2268: 'Das WebDAV-Ziel ist nicht nutzbar: {detail}',
  // The accent color must look like #3b6ef5.
  srv_1c9c8a99006c73fe: 'Die Akzentfarbe muss wie #3b6ef5 aussehen.',
  // The account has already been set up.
  srv_ef560ac46442ad62: 'Das Konto wurde bereits eingerichtet.',
  // The account must be set up first.
  srv_86b531f1833dda66: 'Das Konto muss zuerst eingerichtet werden.',
  // The authentication code is incorrect.
  srv_1f91f70c0f8f35da: 'Der Authentifizierungscode ist falsch.',
  // The background image must be an http(s) URL.
  srv_28004c9379be9be3: 'Das Hintergrundbild muss eine http(s)-URL sein.',
  // The backup could not be received: {detail}
  srv_ad9909a46d053d7b: 'Das Backup konnte nicht empfangen werden: {detail}',
  // The backup does not exist.
  srv_1a3955e489757321: 'Das Backup existiert nicht.',
  // The backup interval must be at least one hour.
  srv_0fd566747f1cd503: 'Das Backup-Intervall muss mindestens eine Stunde betragen.',
  // The backup is damaged (integrity check failed).
  srv_678e4dc426de200a: 'Das Backup ist beschädigt (Integritätsprüfung fehlgeschlagen).',
  // The backup source does not exist.
  srv_24356b769e2a8258: 'Die Backup-Quelle existiert nicht.',
  // The backup was made by a newer version of trisse-home. Update first, then restore it.
  srv_5b20a69f346d31d3:
    'Das Backup stammt von einer neueren trisse-home-Version. Bitte zuerst aktualisieren und dann wiederherstellen.',
  // The bookmark does not exist.
  srv_781f5105c04c646b: 'Das Lesezeichen existiert nicht.',
  // The calendar server is not reachable.
  srv_3e73df212f6df359: 'Der Kalenderserver ist nicht erreichbar.',
  // The calendar server rejected the event: {detail}
  srv_6356e14455c5ce13: 'Der Kalenderserver hat den Termin abgelehnt: {detail}',
  // The certificate warning must be between 1 and 90 days.
  srv_55663fc836edb6a7: 'Die Zertifikatswarnung muss zwischen 1 und 90 Tagen liegen.',
  // The channel could not be reached: {detail}
  srv_d2989a89176a1847: 'Der Kanal ist nicht erreichbar: {detail}',
  // The channel does not exist.
  srv_e97962ce6df48c04: 'Der Kanal existiert nicht.',
  // The channel type must be gotify, webhook or email.
  srv_369ee646cdd39767: 'Die Kanalart muss gotify, webhook oder email sein.',
  // The check URL must start with http:// or https://.
  srv_49713b674a89dd04: 'Die Prüf-URL muss mit http:// oder https:// beginnen.',
  // The custom CSS must be at most 64 KiB.
  srv_b410fbcb84d6e168: 'Das eigene CSS darf höchstens 64 KiB groß sein.',
  // The daily DHL limit is reached; it will be retried later.
  srv_2d07f911dc00e3dc: 'Das tägliche DHL-Limit ist erreicht; es wird später erneut versucht.',
  // The data could not be imported: {detail}
  srv_5715790273b7248a: 'Die Daten konnten nicht importiert werden: {detail}',
  // The departure information is not available right now.
  srv_83a3bc209358c090: 'Die Abfahrtsinformationen sind gerade nicht verfügbar.',
  // The description must be at most 1000 characters long.
  srv_407e3896b9752ec7: 'Die Beschreibung darf höchstens 1000 Zeichen lang sein.',
  // The description must be at most 200 characters long.
  srv_b6d4e2000983a0fe: 'Die Beschreibung darf höchstens 200 Zeichen lang sein.',
  // The disk threshold must be between 50 and 99 %.
  srv_1c6e2427a40c8139: 'Der Speicher-Schwellwert muss zwischen 50 und 99 % liegen.',
  // The due date must look like 2026-12-24.
  srv_8fd5a1f81a5668ce: 'Das Fälligkeitsdatum muss wie 2026-12-24 aussehen.',
  // The event must end after it starts.
  srv_db999c150db53756: 'Der Termin muss nach seinem Beginn enden.',
  // The expiry must be in the future.
  srv_d08d3ed2e05c6dbc: 'Das Ablaufdatum muss in der Zukunft liegen.',
  // The feed URL must start with http:// or https://.
  srv_bbec0c13348226e8: 'Die Feed-URL muss mit http:// oder https:// beginnen.',
  // The feed could not be loaded: {detail}
  srv_4dfea8484179a15f: 'Der Feed konnte nicht geladen werden: {detail}',
  // The feed does not exist.
  srv_5065a34cb1180bb7: 'Der Feed existiert nicht.',
  // The file could not be read (max. 8 MB).
  srv_0371382a6f62f53f: 'Die Datei konnte nicht gelesen werden (max. 8 MB).',
  // The file is larger than 16 MB.
  srv_e7cb725c138134cc: 'Die Datei ist größer als 16 MB.',
  // The file is not a trisse-home backup.
  srv_2dd869ff96b4bc7a: 'Die Datei ist kein trisse-home-Backup.',
  // The file is not valid JSON.
  srv_b8c878135ee917b1: 'Die Datei ist kein gültiges JSON.',
  // The folder does not exist.
  srv_9f6aa27e5380b437: 'Der Ordner existiert nicht.',
  // The folder name must be between 1 and 60 characters long.
  srv_c95a6275d4517459: 'Der Ordnername muss zwischen 1 und 60 Zeichen lang sein.',
  // The forecast warning must be between 1 and 365 days.
  srv_4a2cd38168dace16: 'Die Prognosewarnung muss zwischen 1 und 365 Tagen liegen.',
  // The forecast window must be between 3 and 90 days.
  srv_57fd7519663b39c9: 'Das Prognosefenster muss zwischen 3 und 90 Tagen liegen.',
  // The grace period must be between 0 minutes and 7 days.
  srv_dcd48aa7c1edb891: 'Die Karenzzeit muss zwischen 0 Minuten und 7 Tagen liegen.',
  // The guest does not exist.
  srv_23575e607ab8d86b: 'Der Gast existiert nicht.',
  // The height must be between 1 and 24 rows.
  srv_0d5b9b677899722e: 'Die Höhe muss zwischen 1 und 24 Zeilen liegen.',
  // The icon is too long.
  srv_42e741b2bbfc2695: 'Das Icon ist zu lang.',
  // The identity provider could not be reached: {detail}
  srv_62951ebf93a06287: 'Der Identity Provider ist nicht erreichbar: {detail}',
  // The identity provider did not confirm the sign-in: {detail}
  srv_af05df00dfc4dc1a: 'Der Identity Provider hat die Anmeldung nicht bestätigt: {detail}',
  // The image update check is disabled. Enable it in Settings → Docker updates.
  srv_c9b09323ab5702c6:
    'Die Prüfung auf Image-Updates ist ausgeschaltet. Sie lässt sich unter Einstellungen → Docker-Updates einschalten.',
  // The item does not exist (any more).
  srv_77686a482211f04a: 'Der Eintrag existiert nicht (mehr).',
  // The item does not exist.
  srv_fd944f7eadfdb070: 'Der Eintrag existiert nicht.',
  // The item was changed in another app at the same time. Please try again.
  srv_7faea27f96c82f5f:
    'Der Eintrag wurde gleichzeitig in einer anderen App geändert. Bitte erneut versuchen.',
  // The keyword must consist of 1 to 16 letters or digits.
  srv_d6e09729bc0d1f3b: 'Das Schlüsselwort muss aus 1 bis 16 Buchstaben oder Ziffern bestehen.',
  // The kind must be backup or check.
  srv_0dbc446d2a829ab3: 'Die Art muss backup oder check sein.',
  // The label must be at most 60 characters long.
  srv_64e5ff21727bf802: 'Die Bezeichnung darf höchstens 60 Zeichen lang sein.',
  // The language must be en, de or empty (browser).
  srv_d55a4312396d903f: 'Die Sprache muss en, de oder leer (Browser) sein.',
  // The last tab cannot be deleted.
  srv_fd5437e3a5a60614: 'Der letzte Tab kann nicht gelöscht werden.',
  // The latitude must be between -90 and 90.
  srv_3f60e80ed2350494: 'Der Breitengrad muss zwischen -90 und 90 liegen.',
  // The list URL must start with http:// or https://.
  srv_fb8b7dd6765aca44: 'Die Listen-URL muss mit http:// oder https:// beginnen.',
  // The location must be at most 200 characters long.
  srv_eeec5bbc9fb36341: 'Der Ort darf höchstens 200 Zeichen lang sein.',
  // The longitude must be between -180 and 180.
  srv_39566e8cbb1ec35d: 'Der Längengrad muss zwischen -180 und 180 liegen.',
  // The maintenance window does not exist.
  srv_d6a9caae13404f6e: 'Das Wartungsfenster existiert nicht.',
  // The maximum age must be between 1 hour and 60 days.
  srv_761ff31914abb39c: 'Das Höchstalter muss zwischen 1 Stunde und 60 Tagen liegen.',
  // The name must be 2–32 lowercase letters, digits, dots, dashes or underscores.
  srv_bd0d70cd47d150cb:
    'Der Name muss aus 2–32 Kleinbuchstaben, Ziffern, Punkten, Binde- oder Unterstrichen bestehen.',
  // The name must be at most 40 characters long.
  srv_a81a70289d9d44d1: 'Der Name darf höchstens 40 Zeichen lang sein.',
  // The name must be at most 60 characters long.
  srv_7a44bdf0298cc83f: 'Der Name darf höchstens 60 Zeichen lang sein.',
  // The name must be between 1 and 60 characters long.
  srv_71604d59e2455245: 'Der Name muss zwischen 1 und 60 Zeichen lang sein.',
  // The new order must contain every service exactly once.
  srv_a75b491a5e89aabe: 'Die neue Reihenfolge muss jeden Dienst genau einmal enthalten.',
  // The new order must contain every tab exactly once.
  srv_3364239a85507862: 'Die neue Reihenfolge muss jeden Tab genau einmal enthalten.',
  // The note does not exist.
  srv_f901e92a895881b7: 'Die Notiz existiert nicht.',
  // The note must be at most 100 KiB long.
  srv_4afce48c7ea72d13: 'Die Notiz darf höchstens 100 KiB lang sein.',
  // The ntfy server URL must start with http:// or https://.
  srv_81c618d3f04032a6: 'Die ntfy-Server-URL muss mit http:// oder https:// beginnen.',
  // The ntfy server could not be reached: {detail}
  srv_d77d58a4fac3d674: 'Der ntfy-Server ist nicht erreichbar: {detail}',
  // The off-site backup could not be downloaded: {detail}
  srv_d294e91217ac56a4: 'Das externe Backup konnte nicht heruntergeladen werden: {detail}',
  // The off-site backups could not be listed: {detail}
  srv_3b42f48d84e45ed3: 'Die externen Backups konnten nicht aufgelistet werden: {detail}',
  // The owner account is not a guest.
  srv_a8aed831e812c028: 'Das Besitzerkonto ist kein Gast.',
  // The parcel does not exist.
  srv_f35b41c22f1a21c2: 'Die Sendung existiert nicht.',
  // The passkey does not exist.
  srv_30c99e5e8d1ae003: 'Der Passkey existiert nicht.',
  // The passkey request expired. Please try again.
  srv_754fe73fc055a801: 'Die Passkey-Anfrage ist abgelaufen. Bitte erneut versuchen.',
  // The password is incorrect.
  srv_bed2be84c7fc7252: 'Das Passwort ist falsch.',
  // The password must be at least 12 characters long.
  srv_65fb602b5cafaba0: 'Das Passwort muss mindestens 12 Zeichen lang sein.',
  // The password must be at most 256 characters long.
  srv_91216e078745b556: 'Das Passwort darf höchstens 256 Zeichen lang sein.',
  // The password must not consist of whitespace only.
  srv_be9e30d725fd25b6: 'Das Passwort darf nicht nur aus Leerzeichen bestehen.',
  // The period must be positive and at most 42 days long.
  srv_066ea148faee11ae: 'Der Zeitraum muss positiv und höchstens 42 Tage lang sein.',
  // The port must be between 1 and 65535.
  srv_6566f488a34e8fc3: 'Der Port muss zwischen 1 und 65535 liegen.',
  // The rain radar is only available for Germany.
  srv_d5d2cea1e1d1fa8d: 'Das Regenradar ist nur für Deutschland verfügbar.',
  // The retention must be between 1 and 365 days.
  srv_2367e29c6edb5841: 'Die Aufbewahrungsdauer muss zwischen 1 und 365 Tagen liegen.',
  // The row is out of range.
  srv_088ff184c130f97e: 'Die Zeile liegt außerhalb des Bereichs.',
  // The search query is too long.
  srv_657058c57c5faf46: 'Die Suchanfrage ist zu lang.',
  // The search query must not be empty.
  srv_b828d9f84cec079a: 'Die Suchanfrage darf nicht leer sein.',
  // The server answered {status}.
  srv_63f60053f8b51722: 'Der Server antwortete mit {status}.',
  // The service does not exist.
  srv_df6166e3647fbc04: 'Der Dienst existiert nicht.',
  // The session does not exist (anymore).
  srv_869ef1a450e9ad99: 'Die Sitzung existiert nicht (mehr).',
  // The setup token is invalid.
  srv_d08501b212d94717: 'Das Setup-Token ist ungültig.',
  // The shopping list is not configured yet.
  srv_7d7ee048ea53fd65: 'Die Einkaufsliste ist noch nicht eingerichtet.',
  // The shortcut does not exist.
  srv_2b29e30433c7ad15: 'Das Kürzel existiert nicht.',
  // The sign-in took too long or was started elsewhere. Please try again.
  srv_a9878acf59444198:
    'Die Anmeldung hat zu lange gedauert oder wurde woanders begonnen. Bitte erneut versuchen.',
  // The source must be local, remote or upload.
  srv_7054f9e52efbae1e: 'Die Quelle muss local, remote oder upload sein.',
  // The tab does not exist.
  srv_7e608cff024ed2a8: 'Der Tab existiert nicht.',
  // The tab has too many widgets for a template.
  srv_12af04b41c068c5d: 'Der Tab hat zu viele Widgets für eine Vorlage.',
  // The tab name must be at most 40 characters long.
  srv_e24dab6aca94e3e8: 'Der Tab-Name darf höchstens 40 Zeichen lang sein.',
  // The tab name must not be empty.
  srv_31748b4a01624458: 'Der Tab-Name darf nicht leer sein.',
  // The task list is not reachable: {detail}
  srv_788613837cca8cc6: 'Die Aufgabenliste ist nicht erreichbar: {detail}',
  // The template does not exist.
  srv_4ad756b3998ef1fd: 'Die Vorlage existiert nicht.',
  // The template must be at most 4000 characters long.
  srv_83146ba690ec10f4: 'Die Vorlage darf höchstens 4000 Zeichen lang sein.',
  // The template name must be between 1 and 40 characters long.
  srv_c62fe8c4040fbb11: 'Der Vorlagenname muss zwischen 1 und 40 Zeichen lang sein.',
  // The theme must be system, light or dark.
  srv_d28e94f3f2756a0b: 'Das Design muss system, light oder dark sein.',
  // The title must be at most 200 characters long.
  srv_1fd35c867febde72: 'Der Titel darf höchstens 200 Zeichen lang sein.',
  // The title must be between 1 and 200 characters long.
  srv_f79c303515b1250c: 'Der Titel muss zwischen 1 und 200 Zeichen lang sein.',
  // The to-do does not exist.
  srv_39e64da309876466: 'Das To-do existiert nicht.',
  // The to-do must have between 1 and 300 characters.
  srv_deed27849449d0b5: 'Das To-do muss zwischen 1 und 300 Zeichen haben.',
  // The token does not exist.
  srv_10c62f2d6f8f250e: 'Das Token existiert nicht.',
  // The token is not allowed to do this.
  srv_1399b4677ad575ae: 'Das Token darf das nicht.',
  // The topic may contain letters, digits, - and _ (max. 64 characters).
  srv_218a25cf9751b7dd: 'Das Topic darf Buchstaben, Ziffern, - und _ enthalten (max. 64 Zeichen).',
  // The weather warnings are not available right now.
  srv_314a73e152936a17: 'Die Wetterwarnungen sind gerade nicht verfügbar.',
  // The widget configuration is too large.
  srv_3f4f8e97996c49f7: 'Die Widget-Konfiguration ist zu groß.',
  // The widget configuration must be a JSON object.
  srv_90febdf2f4113f20: 'Die Widget-Konfiguration muss ein JSON-Objekt sein.',
  // The widget does not exist.
  srv_cff3a20eb39a88df: 'Das Widget existiert nicht.',
  // The widget must fit into the 12-column grid.
  srv_fbdc8d1ab5c933c9: 'Das Widget muss in das 12-spaltige Raster passen.',
  // The widget type is invalid.
  srv_a164605f887b18d7: 'Der Widget-Typ ist ungültig.',
  // The width must be between 1 and 12 columns.
  srv_9dfec97d8d360b86: 'Die Breite muss zwischen 1 und 12 Spalten liegen.',
  // There are already 10 channels.
  srv_53e6d03eb73e5456: 'Es gibt bereits 10 Kanäle.',
  // There are already 10 passkeys.
  srv_7c2cacdb15125f01: 'Es gibt bereits 10 Passkeys.',
  // There are already 20 guests.
  srv_07a824f1534514ba: 'Es gibt bereits 20 Gäste.',
  // There are already 50 tokens.
  srv_f6aacc9caebd94cc: 'Es gibt bereits 50 Tokens.',
  // This account is disabled.
  srv_5b702bf7f69da54c: 'Dieses Konto ist deaktiviert.',
  // This account is not allowed to sign in here.
  srv_bd7029b4a21fcba0: 'Dieses Konto darf sich hier nicht anmelden.',
  // This carrier is not supported.
  srv_06ef324c0db01986: 'Dieser Zusteller wird nicht unterstützt.',
  // This is not a WireGuard public key.
  srv_e5774de61b49a80f: 'Das ist kein öffentlicher WireGuard-Schlüssel.',
  // This is not a trisse-home export file.
  srv_d9a79f89c625afc7: 'Das ist keine trisse-home-Exportdatei.',
  // This passkey could not be verified.
  srv_ad0a8083bfc3f124: 'Dieser Passkey konnte nicht überprüft werden.',
  // This passkey is already registered.
  srv_52d3682de07001b1: 'Dieser Passkey ist bereits registriert.',
  // This passkey reported an old signature counter and may have been cloned. Remove it and register it again.
  srv_f30e4d83e431610d:
    'Dieser Passkey hat einen alten Signaturzähler gemeldet und wurde möglicherweise geklont. Entferne ihn und registriere ihn neu.',
  // This widget type is not available.
  srv_63543557e19be101: 'Dieser Widget-Typ ist nicht verfügbar.',
  // Thresholds must be between 0 and 100000 Mbit/s.
  srv_eb4a741553ab9379: 'Schwellenwerte müssen zwischen 0 und 100000 Mbit/s liegen.',
  // Too many failed attempts. Please wait a few minutes.
  srv_bc7172983106e8f0: 'Zu viele Fehlversuche. Bitte warte ein paar Minuten.',
  // Tracking numbers consist of 8 to 35 letters and digits.
  srv_e4cd1895823b2401: 'Sendungsnummern bestehen aus 8 bis 35 Buchstaben und Ziffern.',
  // Two-factor authentication is already enabled.
  srv_a9954572b8880ec0: 'Die Zwei-Faktor-Authentifizierung ist bereits aktiviert.',
  // Two-factor authentication is not enabled.
  srv_326d279028bf3d2f: 'Die Zwei-Faktor-Authentifizierung ist nicht aktiviert.',
  // Unknown alert type {type}.
  srv_fc5b4d698d5d3c6f: 'Unbekannte Benachrichtigungsart {type}.',
  // Unknown search provider.
  srv_dea670a71d2295ca: 'Unbekannte Suchmaschine.',
  // Use a cron expression with five fields, e.g. 0 3 * * * (daily at 03:00).
  srv_564c3b1fecfb1093:
    'Bitte einen Cron-Ausdruck mit fünf Feldern verwenden, z. B. 0 3 * * * (täglich um 03:00).',
  // User name and password (or app password) are required.
  srv_1c73ff4160f4ee94: 'Benutzername und Passwort (oder App-Passwort) sind erforderlich.',
  // Username and app password are required.
  srv_8fe21afcca496dfb: 'Benutzername und App-Passwort sind erforderlich.',
  // Username and password are required.
  srv_ed3a053401c9b56a: 'Benutzername und Passwort sind erforderlich.',
  // Wake-on-LAN is disabled in the demo.
  srv_eceee41d79cdfdb2: 'Wake-on-LAN ist in der Demo deaktiviert.',
  // You are already subscribed to this feed.
  srv_d6a88cffb18a9963: 'Du hast diesen Feed bereits abonniert.',
  // Your session has expired. Please log in again.
  srv_a1f313de1861076a: 'Deine Sitzung ist abgelaufen. Bitte melde dich erneut an.',
  // from and to must be RFC 3339 timestamps.
  srv_ac92ff8f414a83b9: 'from und to müssen RFC-3339-Zeitstempel sein.',
  // id and password are required
  srv_d036d6830bf49051: 'ID und Passwort sind erforderlich',
  // id and type are required
  srv_b672e981cecb99fe: 'ID und Typ sind erforderlich',
  // incomplete credential
  srv_b27eab1b501a12b6: 'unvollständige Zugangsdaten',
  // must be base64url
  srv_0d6f4bdef976a82d: 'muss base64url-kodiert sein',
  // must be positive
  srv_2199d11679859f04: 'muss positiv sein',
  // must not be empty
  srv_705d19e8395a99a3: 'darf nicht leer sein',
  // this is not a valid RSS, Atom or JSON feed
  srv_668a30f80af88b85: 'das ist kein gültiger RSS-, Atom- oder JSON-Feed',
  // token, public id and csrf token are required
  srv_1b3daa6bc9a9f8db: 'Token, öffentliche ID und CSRF-Token sind erforderlich',
  // At most {max} favourite stations.
  srv_64c6ae0369f82879: 'Höchstens {max} Lieblingstankstellen.',
  // Choose E5, E10 or Diesel.
  srv_c2ead036ac61a265: 'Wähle E5, E10 oder Diesel.',
  // Could not load fuel prices: {detail}
  srv_3c850049c1f29436: 'Spritpreise konnten nicht geladen werden: {detail}',
  // Every favourite needs a station ID and a name.
  srv_7c971af06beb10fb: 'Jede Lieblingstankstelle braucht eine Tankstellen-ID und einen Namen.',
  // Fuel prices are not configured yet.
  srv_b8ffa62b5ed7e47e: 'Spritpreise sind noch nicht eingerichtet.',
  // Please choose a location.
  srv_4df94a98701c6c20: 'Bitte wähle einen Ort.',
  // The API key must look like 00000000-0000-0000-0000-000000000002.
  srv_90e0ffa88b068f72: 'Der API-Schlüssel muss wie 00000000-0000-0000-0000-000000000002 aussehen.',
  // The fuel price service is not reachable.
  srv_2542e39dc33a11c6: 'Der Spritpreis-Dienst ist nicht erreichbar.',
  // The place name is too long.
  srv_e4facbe924cb2c30: 'Der Ortsname ist zu lang.',
  // The radius must be between 1 and 25 km.
  srv_ad1457baee3bb1c1: 'Der Umkreis muss zwischen 1 und 25 km liegen.',
  // Docker discovery needs the docker module to be enabled.
  srv_1f3976061912504c: 'Die Docker-Autodiscovery braucht das aktivierte Docker-Modul.',
  // The interval must be between 1 and {max} minutes.
  srv_c2aa1fabd915da16: 'Das Intervall muss zwischen 1 und {max} Minuten liegen.',
  // Unknown merge rule.
  srv_c64f8cf81797f821: 'Unbekannte Abgleichsregel.',
};
