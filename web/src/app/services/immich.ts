import { ImmichApi, ImmichUseCases } from '../../features/immich';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    immich: ImmichUseCases;
  }
}

export const service = defineService('immich', (http) => new ImmichUseCases(new ImmichApi(http)));
