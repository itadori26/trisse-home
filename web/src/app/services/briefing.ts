import { BriefingApi } from '../../features/briefing/adapter/api/briefingApi';
import { BriefingUseCases } from '../../features/briefing/usecase/briefingUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    briefing: BriefingUseCases;
  }
}

export const service = defineService('briefing', (http) => new BriefingUseCases(new BriefingApi(http)));
