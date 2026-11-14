import { FuelApi } from '../../features/fuel/adapter/api/fuelApi';
import { FuelUseCases } from '../../features/fuel/usecase/fuelUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    fuel: FuelUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('fuel', (http) => new FuelUseCases(new FuelApi(http)));
