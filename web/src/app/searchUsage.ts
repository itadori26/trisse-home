import { LocalUsageStore } from '../features/search/adapter/localUsageStore';
import { SearchUsage } from '../features/search/usecase/searchUsage';

/**
 * The search history of this browser (#159), shared by the quick search
 * (opened from the search field) and the settings. Ids of chosen results:
 * "<kind>:<id>" for bookmarks, services, notes and to-dos,
 * "answer:<intent>" for smart answers.
 */
export const searchUsage = new SearchUsage(new LocalUsageStore());

export const answerTarget = (intent: string) => `answer:${intent}`;
