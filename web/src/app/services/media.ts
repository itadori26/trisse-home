import { MediaApi, MediaUseCases } from '../../features/media';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    media: MediaUseCases;
  }
}

export const service = defineService('media', (http) => new MediaUseCases(new MediaApi(http)));
