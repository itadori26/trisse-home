import type { en } from './en';

/** German UI texts of the AI summaries settings; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'settings.section.summary': 'KI-Zusammenfassungen',
  'summary.hint':
    'Notizen und Feed-Artikel mit einem lokalen Ollama zusammenfassen. Nichts verlässt dein Heimnetz; solange nichts eingerichtet ist, bleibt die Funktion verborgen.',
  'summary.url': 'Ollama-Adresse (Heimnetz)',
  'summary.load_models': 'Modelle laden',
  'summary.model': 'Modell',
  'summary.models_found': 'Gefundene Modelle: {n}.',
  'summary.no_models': 'Keine Modelle installiert; lade zuerst eines, z. B. ollama pull llama3.2:3b.',
  'summary.max_input': 'Höchstens {n} Zeichen zusammenfassen',
  'summary.save': 'Speichern',
  'summary.saved': 'Verbunden und gespeichert. Notizen und Feed-Einträge zeigen jetzt „Zusammenfassen“.',
  'summary.disable': 'Ausschalten',
  'summary.disabled': 'Ausgeschaltet; gespeicherte Zusammenfassungen gelöscht.',
};
