// Imported from the files, not the feature index: the index also exports the
// settings component, which would otherwise end up in the initial bundle (#230).
import { HolidaysApi } from '../../features/holidays/adapter/api/holidaysApi';
import { HolidaysUseCases } from '../../features/holidays/usecase/holidaysUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    holidays: HolidaysUseCases;
  }
}

export const service = defineService('holidays', (http) => new HolidaysUseCases(new HolidaysApi(http)));
