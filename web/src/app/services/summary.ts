import { SummaryApi } from '../../features/summary/adapter/api/summaryApi';
import { SummaryUseCases } from '../../features/summary/usecase/summaryUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    summary: SummaryUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('summary', (http) => new SummaryUseCases(new SummaryApi(http)));
