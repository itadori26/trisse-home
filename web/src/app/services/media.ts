import { MediaApi } from '../../features/media/adapter/api/mediaApi';
import { MediaUseCases } from '../../features/media/usecase/mediaUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    media: MediaUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('media', (http) => new MediaUseCases(new MediaApi(http)));
