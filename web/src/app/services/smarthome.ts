import { SmarthomeApi } from '../../features/smarthome/adapter/api/smarthomeApi';
import { SmarthomeUseCases } from '../../features/smarthome/usecase/smarthomeUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    smarthome: SmarthomeUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('smarthome', (http) => new SmarthomeUseCases(new SmarthomeApi(http)));
