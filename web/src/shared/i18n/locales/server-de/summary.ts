import type { ServerMessageKey } from '../../../api-client/generated/server-messages/summary';

/**
 * German translations of the backend error messages of `summary` (#75). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose a model of your Ollama server.
  srv_8934ba71a8244963: 'Wähle ein Modell deines Ollama-Servers.',
  // Ollama could not summarise the text: {detail}
  srv_49d6ef3b7320de2a: 'Ollama konnte den Text nicht zusammenfassen: {detail}',
  // Ollama returned an empty summary.
  srv_d3de1918e0e0bf92: 'Ollama hat eine leere Zusammenfassung geliefert.',
  // Only an Ollama in your home network is allowed, e.g. http://192.168.1.10:11434 or http://ollama.lan:11434.
  srv_1dee19acecab138c:
    'Nur ein Ollama in deinem Heimnetz ist erlaubt, z. B. http://192.168.1.10:11434 oder http://ollama.lan:11434.',
  // Summaries are not set up; add your Ollama server in the settings.
  srv_2038bcfc12517b92:
    'Zusammenfassungen sind nicht eingerichtet; trage deinen Ollama-Server in den Einstellungen ein.',
  // The article could not be loaded: {detail}
  srv_d741b64ac0a7b38d: 'Der Artikel konnte nicht geladen werden: {detail}',
  // The input limit must be between 1,000 and 50,000 characters.
  srv_d000a1cffb2b61b2: 'Die Eingabegrenze muss zwischen 1.000 und 50.000 Zeichen liegen.',
  // The model {model} is not installed on this Ollama server.
  srv_913837b8f399a5a1: 'Das Modell {model} ist auf diesem Ollama-Server nicht installiert.',
  // There is no text to summarise.
  srv_a71ddd931e1d1858: 'Es gibt keinen Text zum Zusammenfassen.',
  // Ask a question after the question mark.
  srv_086f1e1c1471afa7: 'Stell nach dem Fragezeichen eine Frage.',
  // The question must be at most 300 characters long.
  srv_e0be8094bf6d2de5: 'Die Frage darf höchstens 300 Zeichen lang sein.',
};
