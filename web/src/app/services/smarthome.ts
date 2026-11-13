import { SmarthomeApi, SmarthomeUseCases } from '../../features/smarthome';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    smarthome: SmarthomeUseCases;
  }
}

export const service = defineService('smarthome', (http) => new SmarthomeUseCases(new SmarthomeApi(http)));
