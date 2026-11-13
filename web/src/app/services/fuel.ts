import { FuelApi, FuelUseCases } from '../../features/fuel';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    fuel: FuelUseCases;
  }
}

export const service = defineService('fuel', (http) => new FuelUseCases(new FuelApi(http)));
