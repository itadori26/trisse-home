import { ImmichApi } from '../../features/immich/adapter/api/immichApi';
import { ImmichUseCases } from '../../features/immich/usecase/immichUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    immich: ImmichUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('immich', (http) => new ImmichUseCases(new ImmichApi(http)));
