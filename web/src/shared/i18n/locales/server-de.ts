import type { ServerMessageKey } from '../../api-client/generated/server-messages';

/**
 * German translations of the backend error messages (#102). Keys come from
 * the generated catalogue; the type makes a missing translation a compile
 * error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A bookmark can have at most 10 tags.
  srv_5aca9a96f330: 'Ein Lesezeichen kann höchstens 10 Tags haben.',
  // A guest with this name already exists.
  srv_2ee9dea7e143: 'Es gibt bereits einen Gast mit diesem Namen.',
  // A note needs a title or some text.
  srv_1c46ab642663: 'Eine Notiz braucht einen Titel oder etwas Text.',
  // A shortcut with this keyword already exists.
  srv_fdb1a335cf13: 'Es gibt bereits ein Kürzel mit diesem Schlüsselwort.',
  // Actions are not enabled for this container (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).
  srv_dda25b75cb78:
    'Aktionen sind für diesen Container nicht freigeschaltet (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).',
  // AdGuard Home is not configured yet.
  srv_5609c316813a: 'AdGuard Home ist noch nicht eingerichtet.',
  // AdGuard Home is not reachable.
  srv_f091d8c36fae: 'AdGuard Home ist nicht erreichbar.',
  // At most 10 calendars are supported.
  srv_738a14a7fbb1: 'Es werden höchstens 10 Kalender unterstützt.',
  // At most 20 rules are supported.
  srv_7db3413fcbca: 'Es werden höchstens 20 Regeln unterstützt.',
  // At most 30 parcels can be tracked.
  srv_505e970ff017: 'Es können höchstens 30 Sendungen verfolgt werden.',
  // At most 50 templates can be stored.
  srv_02c0cbbf2137: 'Es können höchstens 50 Vorlagen gespeichert werden.',
  // At most {max} bookmarks can be imported at once.
  srv_4fa8973eb3b0: 'Es können höchstens {max} Lesezeichen auf einmal importiert werden.',
  // Between 1 and 365 backups can be kept.
  srv_180345ed5de4: 'Es können zwischen 1 und 365 Backups aufbewahrt werden.',
  // Choose at least one scope.
  srv_cb12f8b742e6: 'Bitte mindestens einen Bereich wählen.',
  // Choose between 0 (never) and 3650 days.
  srv_dd2711750d93: 'Bitte zwischen 0 (nie) und 3650 Tagen wählen.',
  // Colors must look like #1a2b3c.
  srv_7346d81f035f: 'Farben müssen wie #1a2b3c aussehen.',
  // Could not connect to AdGuard Home: {detail}
  srv_81950d96ea70: 'Keine Verbindung zu AdGuard Home: {detail}',
  // Could not connect to Nextcloud: {detail}
  srv_2242f13e4baf: 'Keine Verbindung zu Nextcloud: {detail}',
  // Could not connect to the FRITZ!Box: {detail}
  srv_d6bef69d57dd: 'Keine Verbindung zur FRITZ!Box: {detail}',
  // Could not load the calendar: {detail}
  srv_49797fb7c917: 'Der Kalender konnte nicht geladen werden: {detail}',
  // Could not load the waste calendar: {detail}
  srv_cf0a135992e7: 'Der Abfallkalender konnte nicht geladen werden: {detail}',
  // Could not read the task list: {detail}
  srv_ebf642c7a8be: 'Die Aufgabenliste konnte nicht gelesen werden: {detail}',
  // DHL does not know this tracking number (yet).
  srv_6ad773f8f8b7: 'DHL kennt diese Sendungsnummer (noch) nicht.',
  // DHL rejected the API key.
  srv_fbfc8b373507: 'DHL hat den API-Schlüssel abgelehnt.',
  // Delivered parcels can be kept for 0 to 30 days.
  srv_d33b345a83fa: 'Zugestellte Sendungen können 0 bis 30 Tage angezeigt werden.',
  // Enter your authentication code to finish signing in.
  srv_1a32b19e0062: 'Gib deinen Authentifizierungscode ein, um die Anmeldung abzuschließen.',
  // Events can be at most one year long.
  srv_7af7334d30be: 'Termine können höchstens ein Jahr lang sein.',
  // Every calendar URL must start with http:// or https://.
  srv_67287d89a43a: 'Jede Kalender-URL muss mit http:// oder https:// beginnen.',
  // Every rule needs a match text of at most 40 characters.
  srv_e2f87ec8f536: 'Jede Regel braucht einen Suchtext mit höchstens 40 Zeichen.',
  // Export version {version} is not supported by this version of trisse-home.
  srv_72596ceac051: 'Exportversion {version} wird von dieser trisse-home-Version nicht unterstützt.',
  // Guests can only view this dashboard.
  srv_4e8139043900: 'Gäste können dieses Dashboard nur ansehen.',
  // Host metrics are not available. Are /proc and /sys mounted?
  srv_8f4bb404d952: 'Host-Messwerte sind nicht verfügbar. Sind /proc und /sys eingebunden?',
  // Icons must be at most 8 characters long.
  srv_347778ecbfa9: 'Icons dürfen höchstens 8 Zeichen lang sein.',
  // Invalid scope {scope}: use module:read or module:write.
  srv_ed38267b6348: 'Ungültiger Bereich {scope}: bitte modul:read oder modul:write verwenden.',
  // Items must be between 1 and 200 characters long.
  srv_a23b7aa31f22: 'Einträge müssen zwischen 1 und 200 Zeichen lang sein.',
  // Labels must be at most 40 characters long.
  srv_6f5f3ffdc73b: 'Bezeichnungen dürfen höchstens 40 Zeichen lang sein.',
  // Latitude and longitude are required.
  srv_9d96c205040c: 'Breiten- und Längengrad sind erforderlich.',
  // Login failed: {detail}
  srv_8742c300e84a: 'Anmeldung fehlgeschlagen: {detail}',
  // Missing or invalid CSRF token.
  srv_e876378327a3: 'CSRF-Token fehlt oder ist ungültig.',
  // Nextcloud is not configured yet.
  srv_b24a7c98161b: 'Nextcloud ist noch nicht eingerichtet.',
  // No MAC address is stored for this service.
  srv_2916787131c0: 'Für diesen Dienst ist keine MAC-Adresse hinterlegt.',
  // No WireGuard export found. Run scripts/wireguard-export.sh on the host (see README).
  srv_f4b1a0014fbc:
    'Kein WireGuard-Export gefunden. Führe scripts/wireguard-export.sh auf dem Host aus (siehe README).',
  // No account has been set up yet.
  srv_97490302ba2a: 'Es wurde noch kein Konto eingerichtet.',
  // No bookmarks found. Please export your bookmarks as HTML from the browser.
  srv_97ecaeb3a34e: 'Keine Lesezeichen gefunden. Bitte exportiere deine Lesezeichen im Browser als HTML.',
  // No calendar is configured yet.
  srv_71a43ecc2251: 'Es ist noch kein Kalender eingerichtet.',
  // No waste calendar is configured yet.
  srv_b56a0a7e5341: 'Es ist noch kein Abfallkalender eingerichtet.',
  // Only the owner can do this.
  srv_010dd173e73d: 'Das kann nur der Besitzer.',
  // Passkeys need HTTPS and PUBLIC_URL (see the README).
  srv_381af756d6e9: 'Passkeys brauchen HTTPS und PUBLIC_URL (siehe README).',
  // Password login is disabled. Use single sign-on.
  srv_dfd6230a11b1: 'Die Anmeldung mit Passwort ist deaktiviert. Bitte Single Sign-on verwenden.',
  // Please add your DHL API key in the settings first.
  srv_8e51d0499ac7: 'Bitte zuerst deinen DHL-API-Schlüssel in den Einstellungen hinterlegen.',
  // Please choose a stop in the widget settings.
  srv_63efa2ecd7e4: 'Bitte wähle in den Widget-Einstellungen eine Haltestelle.',
  // Please choose one of the configured calendars.
  srv_abacb0b52864: 'Bitte wähle einen der eingerichteten Kalender.',
  // Please enter a valid start and end.
  srv_9642184dcf74: 'Bitte gib einen gültigen Beginn und ein gültiges Ende ein.',
  // Please enter and save a topic first.
  srv_83bd0750c826: 'Bitte zuerst ein Topic eingeben und speichern.',
  // Please enter at least one calendar URL.
  srv_18810d4dff85: 'Bitte mindestens eine Kalender-URL eingeben.',
  // Please enter between 2 and 100 characters.
  srv_6fc922ccabdf: 'Bitte zwischen 2 und 100 Zeichen eingeben.',
  // Please enter the code from your authenticator app or a recovery code.
  srv_3cea16af2934: 'Bitte gib den Code aus deiner Authenticator-App oder einen Wiederherstellungscode ein.',
  // Please enter the password of the FRITZ!Box user.
  srv_ca4bac17e30c: 'Bitte gib das Passwort des FRITZ!Box-Benutzers ein.',
  // Session not found.
  srv_0d8b0cf2daf3: 'Sitzung nicht gefunden.',
  // Single sign-on is not configured.
  srv_2402f193e4e9: 'Single Sign-on ist nicht eingerichtet.',
  // Start the two-factor setup first.
  srv_3d224aee3fde: 'Bitte zuerst die Einrichtung der Zwei-Faktor-Authentifizierung starten.',
  // Storing credentials is disabled in the demo.
  srv_552706103662: 'Das Speichern von Zugangsdaten ist in der Demo deaktiviert.',
  // Tags may only contain letters, digits, - and _ (max. 30 characters).
  srv_b74ac24eda46: 'Tags dürfen nur Buchstaben, Ziffern, - und _ enthalten (max. 30 Zeichen).',
  // The FRITZ!Box is not configured yet.
  srv_baa8e544f8db: 'Die FRITZ!Box ist noch nicht eingerichtet.',
  // The ICS URL must start with https://, http:// or webcal://.
  srv_50c914c9ae43: 'Die ICS-URL muss mit https://, http:// oder webcal:// beginnen.',
  // The URL must contain {query}.
  srv_e6597f337bc6: 'Die URL muss {query} enthalten.',
  // The URL must look like http://192.168.1.2:3000.
  srv_ef8ffb559478: 'Die URL muss wie http://192.168.1.2:3000 aussehen.',
  // The URL must look like http://192.168.178.1:49000.
  srv_e2dac68431ef: 'Die URL muss wie http://192.168.178.1:49000 aussehen.',
  // The URL must look like https://cloud.example.com.
  srv_1ba57406acfa: 'Die URL muss wie https://cloud.example.com aussehen.',
  // The URL must start with http:// or https://.
  srv_41b41b9013da: 'Die URL muss mit http:// oder https:// beginnen.',
  // The Wake-on-LAN relay did not answer. Is COMPOSE_PROFILES=wol set (see README)?
  srv_8e2f7f76deb1:
    'Das Wake-on-LAN-Relay hat nicht geantwortet. Ist COMPOSE_PROFILES=wol gesetzt (siehe README)?',
  // The accent color must look like #3b6ef5.
  srv_41d7237bec0a: 'Die Akzentfarbe muss wie #3b6ef5 aussehen.',
  // The account has already been set up.
  srv_05cc7a9e1c5b: 'Das Konto wurde bereits eingerichtet.',
  // The account must be set up first.
  srv_13183aa91ba9: 'Das Konto muss zuerst eingerichtet werden.',
  // The authentication code is incorrect.
  srv_a2f38092e89b: 'Der Authentifizierungscode ist falsch.',
  // The background image must be an http(s) URL.
  srv_dfbc7a48eda5: 'Das Hintergrundbild muss eine http(s)-URL sein.',
  // The backup does not exist.
  srv_cbeb51779fd8: 'Das Backup existiert nicht.',
  // The backup interval must be at least one hour.
  srv_7796988474f7: 'Das Backup-Intervall muss mindestens eine Stunde betragen.',
  // The backup source does not exist.
  srv_88e8fa469ffe: 'Die Backup-Quelle existiert nicht.',
  // The bookmark does not exist.
  srv_031256ee8dc1: 'Das Lesezeichen existiert nicht.',
  // The calendar server is not reachable.
  srv_0d2a8f4804ec: 'Der Kalenderserver ist nicht erreichbar.',
  // The calendar server rejected the event: {detail}
  srv_deb31bbc3fc1: 'Der Kalenderserver hat den Termin abgelehnt: {detail}',
  // The certificate warning must be between 1 and 90 days.
  srv_9d502f787b5a: 'Die Zertifikatswarnung muss zwischen 1 und 90 Tagen liegen.',
  // The check URL must start with http:// or https://.
  srv_ffbe50d5f621: 'Die Prüf-URL muss mit http:// oder https:// beginnen.',
  // The custom CSS must be at most 64 KiB.
  srv_9292274232bc: 'Das eigene CSS darf höchstens 64 KiB groß sein.',
  // The daily DHL limit is reached; it will be retried later.
  srv_ce434fa9181b: 'Das tägliche DHL-Limit ist erreicht; es wird später erneut versucht.',
  // The data could not be imported: {detail}
  srv_d18a345d8740: 'Die Daten konnten nicht importiert werden: {detail}',
  // The departure information is not available right now.
  srv_65eb6e88294d: 'Die Abfahrtsinformationen sind gerade nicht verfügbar.',
  // The description must be at most 1000 characters long.
  srv_e5556b8fdf38: 'Die Beschreibung darf höchstens 1000 Zeichen lang sein.',
  // The description must be at most 200 characters long.
  srv_3ce0ba29241f: 'Die Beschreibung darf höchstens 200 Zeichen lang sein.',
  // The disk threshold must be between 50 and 99 %.
  srv_be20661d610c: 'Der Speicher-Schwellwert muss zwischen 50 und 99 % liegen.',
  // The due date must look like 2026-12-24.
  srv_9e526967c18f: 'Das Fälligkeitsdatum muss wie 2026-12-24 aussehen.',
  // The event must end after it starts.
  srv_aaea491b2cb7: 'Der Termin muss nach seinem Beginn enden.',
  // The expiry must be in the future.
  srv_c1f43595726e: 'Das Ablaufdatum muss in der Zukunft liegen.',
  // The feed URL must start with http:// or https://.
  srv_0adafc8c0e5b: 'Die Feed-URL muss mit http:// oder https:// beginnen.',
  // The feed could not be loaded: {detail}
  srv_78440c3e89b1: 'Der Feed konnte nicht geladen werden: {detail}',
  // The feed does not exist.
  srv_c45d14991496: 'Der Feed existiert nicht.',
  // The file could not be read (max. 8 MB).
  srv_df0faae1346c: 'Die Datei konnte nicht gelesen werden (max. 8 MB).',
  // The file is larger than 16 MB.
  srv_0dfcc20d75e9: 'Die Datei ist größer als 16 MB.',
  // The file is not valid JSON.
  srv_6ce00fa0ea10: 'Die Datei ist kein gültiges JSON.',
  // The folder does not exist.
  srv_f2bff02cb2d0: 'Der Ordner existiert nicht.',
  // The folder name must be between 1 and 60 characters long.
  srv_f4a85cd55a33: 'Der Ordnername muss zwischen 1 und 60 Zeichen lang sein.',
  // The guest does not exist.
  srv_92e09fdecdc6: 'Der Gast existiert nicht.',
  // The height must be between 1 and 24 rows.
  srv_1fd7387d21a8: 'Die Höhe muss zwischen 1 und 24 Zeilen liegen.',
  // The icon is too long.
  srv_6e3950d3fda8: 'Das Icon ist zu lang.',
  // The identity provider could not be reached: {detail}
  srv_5b54195e111b: 'Der Identity Provider ist nicht erreichbar: {detail}',
  // The identity provider did not confirm the sign-in: {detail}
  srv_2f221a3f06bb: 'Der Identity Provider hat die Anmeldung nicht bestätigt: {detail}',
  // The item does not exist (any more).
  srv_109c59e28430: 'Der Eintrag existiert nicht (mehr).',
  // The item does not exist.
  srv_0781fd56e132: 'Der Eintrag existiert nicht.',
  // The item was changed in another app at the same time. Please try again.
  srv_b40e38b52770: 'Der Eintrag wurde gleichzeitig in einer anderen App geändert. Bitte erneut versuchen.',
  // The keyword must consist of 1 to 16 letters or digits.
  srv_4d28ae808efc: 'Das Schlüsselwort muss aus 1 bis 16 Buchstaben oder Ziffern bestehen.',
  // The label must be at most 60 characters long.
  srv_2c01d1f58d21: 'Die Bezeichnung darf höchstens 60 Zeichen lang sein.',
  // The language must be en, de or empty (browser).
  srv_147d597d8c16: 'Die Sprache muss en, de oder leer (Browser) sein.',
  // The last tab cannot be deleted.
  srv_7cb136f70661: 'Der letzte Tab kann nicht gelöscht werden.',
  // The latitude must be between -90 and 90.
  srv_fdd8193a3a91: 'Der Breitengrad muss zwischen -90 und 90 liegen.',
  // The list URL must start with http:// or https://.
  srv_8ccad1654f79: 'Die Listen-URL muss mit http:// oder https:// beginnen.',
  // The location must be at most 200 characters long.
  srv_2aad9ee89232: 'Der Ort darf höchstens 200 Zeichen lang sein.',
  // The longitude must be between -180 and 180.
  srv_1ab529cfa267: 'Der Längengrad muss zwischen -180 und 180 liegen.',
  // The maximum age must be between 1 hour and 60 days.
  srv_1bc5e13a457e: 'Das Höchstalter muss zwischen 1 Stunde und 60 Tagen liegen.',
  // The name must be 2–32 lowercase letters, digits, dots, dashes or underscores.
  srv_47659892367a:
    'Der Name muss aus 2–32 Kleinbuchstaben, Ziffern, Punkten, Binde- oder Unterstrichen bestehen.',
  // The name must be at most 40 characters long.
  srv_aa454731e20c: 'Der Name darf höchstens 40 Zeichen lang sein.',
  // The name must be at most 60 characters long.
  srv_0654a1f6ccc2: 'Der Name darf höchstens 60 Zeichen lang sein.',
  // The name must be between 1 and 60 characters long.
  srv_35b693b142b5: 'Der Name muss zwischen 1 und 60 Zeichen lang sein.',
  // The new order must contain every service exactly once.
  srv_495bc7eee4df: 'Die neue Reihenfolge muss jeden Dienst genau einmal enthalten.',
  // The new order must contain every tab exactly once.
  srv_8b7d74b934a2: 'Die neue Reihenfolge muss jeden Tab genau einmal enthalten.',
  // The note does not exist.
  srv_7dd45be91d5c: 'Die Notiz existiert nicht.',
  // The note must be at most 100 KiB long.
  srv_98e528e5ce6e: 'Die Notiz darf höchstens 100 KiB lang sein.',
  // The ntfy server URL must start with http:// or https://.
  srv_f2e40d2033f6: 'Die ntfy-Server-URL muss mit http:// oder https:// beginnen.',
  // The ntfy server could not be reached: {detail}
  srv_e65306dc99a9: 'Der ntfy-Server ist nicht erreichbar: {detail}',
  // The owner account is not a guest.
  srv_6a915d5fa7cb: 'Das Besitzerkonto ist kein Gast.',
  // The parcel does not exist.
  srv_742f3c5e2f06: 'Die Sendung existiert nicht.',
  // The passkey does not exist.
  srv_3712ff041b14: 'Der Passkey existiert nicht.',
  // The passkey request expired. Please try again.
  srv_a95fcec2ba0c: 'Die Passkey-Anfrage ist abgelaufen. Bitte erneut versuchen.',
  // The password is incorrect.
  srv_96bfe6916d71: 'Das Passwort ist falsch.',
  // The password must be at least 12 characters long.
  srv_4c2439788838: 'Das Passwort muss mindestens 12 Zeichen lang sein.',
  // The password must be at most 256 characters long.
  srv_a6013a53c206: 'Das Passwort darf höchstens 256 Zeichen lang sein.',
  // The password must not consist of whitespace only.
  srv_d858457394af: 'Das Passwort darf nicht nur aus Leerzeichen bestehen.',
  // The period must be positive and at most 42 days long.
  srv_e70d096e47a4: 'Der Zeitraum muss positiv und höchstens 42 Tage lang sein.',
  // The rain radar is only available for Germany.
  srv_2095402d98bf: 'Das Regenradar ist nur für Deutschland verfügbar.',
  // The row is out of range.
  srv_89af7ad4573a: 'Die Zeile liegt außerhalb des Bereichs.',
  // The search query is too long.
  srv_e79364d9dd59: 'Die Suchanfrage ist zu lang.',
  // The search query must not be empty.
  srv_558ed0c106eb: 'Die Suchanfrage darf nicht leer sein.',
  // The server answered {status}.
  srv_80ae49e23696: 'Der Server antwortete mit {status}.',
  // The service does not exist.
  srv_241c3d9ab750: 'Der Dienst existiert nicht.',
  // The session does not exist (anymore).
  srv_d46d274229f3: 'Die Sitzung existiert nicht (mehr).',
  // The setup token is invalid.
  srv_46c9b874271d: 'Das Setup-Token ist ungültig.',
  // The shopping list is not configured yet.
  srv_5d7423c4c718: 'Die Einkaufsliste ist noch nicht eingerichtet.',
  // The shortcut does not exist.
  srv_8852840dc7fe: 'Das Kürzel existiert nicht.',
  // The sign-in took too long or was started elsewhere. Please try again.
  srv_da7b0b243a77:
    'Die Anmeldung hat zu lange gedauert oder wurde woanders begonnen. Bitte erneut versuchen.',
  // The tab does not exist.
  srv_35e3061f6ff1: 'Der Tab existiert nicht.',
  // The tab has too many widgets for a template.
  srv_469c1450c44a: 'Der Tab hat zu viele Widgets für eine Vorlage.',
  // The tab name must be at most 40 characters long.
  srv_e25f5f09f16a: 'Der Tab-Name darf höchstens 40 Zeichen lang sein.',
  // The tab name must not be empty.
  srv_46177f46e416: 'Der Tab-Name darf nicht leer sein.',
  // The task list is not reachable: {detail}
  srv_209a17a44dc2: 'Die Aufgabenliste ist nicht erreichbar: {detail}',
  // The template does not exist.
  srv_577c9d2b8734: 'Die Vorlage existiert nicht.',
  // The template name must be between 1 and 40 characters long.
  srv_37bdc6c80e82: 'Der Vorlagenname muss zwischen 1 und 40 Zeichen lang sein.',
  // The theme must be system, light or dark.
  srv_4695e31aacdc: 'Das Design muss system, light oder dark sein.',
  // The title must be at most 200 characters long.
  srv_3e573bed43e3: 'Der Titel darf höchstens 200 Zeichen lang sein.',
  // The title must be between 1 and 200 characters long.
  srv_59fb52a183d5: 'Der Titel muss zwischen 1 und 200 Zeichen lang sein.',
  // The to-do does not exist.
  srv_a1aabee7b460: 'Das To-do existiert nicht.',
  // The to-do must have between 1 and 300 characters.
  srv_a89e460a941c: 'Das To-do muss zwischen 1 und 300 Zeichen haben.',
  // The token does not exist.
  srv_1a55e780d794: 'Das Token existiert nicht.',
  // The token is not allowed to do this.
  srv_1d46d9c03e72: 'Das Token darf das nicht.',
  // The topic may contain letters, digits, - and _ (max. 64 characters).
  srv_f916bf4925b9: 'Das Topic darf Buchstaben, Ziffern, - und _ enthalten (max. 64 Zeichen).',
  // The weather warnings are not available right now.
  srv_07680ee9c73a: 'Die Wetterwarnungen sind gerade nicht verfügbar.',
  // The widget configuration is too large.
  srv_fca003dd121a: 'Die Widget-Konfiguration ist zu groß.',
  // The widget configuration must be a JSON object.
  srv_49fb541deacf: 'Die Widget-Konfiguration muss ein JSON-Objekt sein.',
  // The widget does not exist.
  srv_8a4d3cb33113: 'Das Widget existiert nicht.',
  // The widget must fit into the 12-column grid.
  srv_10e23fb2ded7: 'Das Widget muss in das 12-spaltige Raster passen.',
  // The widget type is invalid.
  srv_568db9fe41b7: 'Der Widget-Typ ist ungültig.',
  // The width must be between 1 and 12 columns.
  srv_c49d63e6ac43: 'Die Breite muss zwischen 1 und 12 Spalten liegen.',
  // There are already 10 passkeys.
  srv_07bb2b659f1a: 'Es gibt bereits 10 Passkeys.',
  // There are already 20 guests.
  srv_fc3fea99e8ec: 'Es gibt bereits 20 Gäste.',
  // There are already 50 tokens.
  srv_ac37d9aac451: 'Es gibt bereits 50 Tokens.',
  // This account is disabled.
  srv_d1294d098e7f: 'Dieses Konto ist deaktiviert.',
  // This account is not allowed to sign in here.
  srv_2721f1002d7e: 'Dieses Konto darf sich hier nicht anmelden.',
  // This carrier is not supported.
  srv_8eb560f361eb: 'Dieser Zusteller wird nicht unterstützt.',
  // This is not a WireGuard public key.
  srv_f25420429765: 'Das ist kein öffentlicher WireGuard-Schlüssel.',
  // This is not a trisse-home export file.
  srv_2ae844e97b0e: 'Das ist keine trisse-home-Exportdatei.',
  // This passkey could not be verified.
  srv_e9a89b21aa46: 'Dieser Passkey konnte nicht überprüft werden.',
  // This passkey is already registered.
  srv_81e308e425eb: 'Dieser Passkey ist bereits registriert.',
  // This passkey reported an old signature counter and may have been cloned. Remove it and register it again.
  srv_f04dfe2ff7fc:
    'Dieser Passkey hat einen alten Signaturzähler gemeldet und wurde möglicherweise geklont. Entferne ihn und registriere ihn neu.',
  // This widget type is not available.
  srv_7e9a52b5c71f: 'Dieser Widget-Typ ist nicht verfügbar.',
  // Too many failed attempts. Please wait a few minutes.
  srv_56f55bf8cb80: 'Zu viele Fehlversuche. Bitte warte ein paar Minuten.',
  // Tracking numbers consist of 8 to 35 letters and digits.
  srv_13e00f02a16e: 'Sendungsnummern bestehen aus 8 bis 35 Buchstaben und Ziffern.',
  // Two-factor authentication is already enabled.
  srv_ac0b7deafa1a: 'Die Zwei-Faktor-Authentifizierung ist bereits aktiviert.',
  // Two-factor authentication is not enabled.
  srv_881118d71b9a: 'Die Zwei-Faktor-Authentifizierung ist nicht aktiviert.',
  // Unknown search provider.
  srv_6ca556a76910: 'Unbekannte Suchmaschine.',
  // Username and app password are required.
  srv_f2ce98955a14: 'Benutzername und App-Passwort sind erforderlich.',
  // Username and password are required.
  srv_b32a800ba7ae: 'Benutzername und Passwort sind erforderlich.',
  // Wake-on-LAN is disabled in the demo.
  srv_87030fe199db: 'Wake-on-LAN ist in der Demo deaktiviert.',
  // You are already subscribed to this feed.
  srv_428d570e725c: 'Du hast diesen Feed bereits abonniert.',
  // Your session has expired. Please log in again.
  srv_05c0eb4cd3b3: 'Deine Sitzung ist abgelaufen. Bitte melde dich erneut an.',
  // from and to must be RFC 3339 timestamps.
  srv_c714f8da642c: 'from und to müssen RFC-3339-Zeitstempel sein.',
  // id and password are required
  srv_78899e632d50: 'ID und Passwort sind erforderlich',
  // id and type are required
  srv_cbbae08e3454: 'ID und Typ sind erforderlich',
  // incomplete credential
  srv_7e878451f1e7: 'unvollständige Zugangsdaten',
  // must be base64url
  srv_c172042b0a1e: 'muss base64url-kodiert sein',
  // must be positive
  srv_c84d22d5069e: 'muss positiv sein',
  // must not be empty
  srv_61507211d861: 'darf nicht leer sein',
  // this is not a valid RSS, Atom or JSON feed
  srv_8f1923823585: 'das ist kein gültiger RSS-, Atom- oder JSON-Feed',
  // token, public id and csrf token are required
  srv_a52a49787408: 'Token, öffentliche ID und CSRF-Token sind erforderlich',
};
