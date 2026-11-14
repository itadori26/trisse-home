import { ArrApi } from '../../features/arr/adapter/api/arrApi';
import { ArrUseCases } from '../../features/arr/usecase/arrUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    arr: ArrUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('arr', (http) => new ArrUseCases(new ArrApi(http)));
