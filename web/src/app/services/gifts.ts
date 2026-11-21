import { GiftsApi } from '../../features/gifts/adapter/api/giftsApi';
import { GiftsUseCases } from '../../features/gifts/usecase/giftsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    gifts: GiftsUseCases;
  }
}

export const service = defineService('gifts', (http) => new GiftsUseCases(new GiftsApi(http)));
