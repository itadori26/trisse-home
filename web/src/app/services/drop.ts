import { DropApi } from '../../features/drop/adapter/api/dropApi';
import { DropUseCases } from '../../features/drop/usecase/dropUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    drop: DropUseCases;
  }
}

export const service = defineService('drop', (http) => new DropUseCases(new DropApi(http)));
