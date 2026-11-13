import { ArrApi, ArrUseCases } from '../../features/arr';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    arr: ArrUseCases;
  }
}

export const service = defineService('arr', (http) => new ArrUseCases(new ArrApi(http)));
