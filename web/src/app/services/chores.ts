import { ChoresApi } from '../../features/chores/adapter/api/choresApi';
import { ChoresUseCases } from '../../features/chores/usecase/choresUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    chores: ChoresUseCases;
  }
}

export const service = defineService('chores', (http) => new ChoresUseCases(new ChoresApi(http)));
