// Imported from the files, not the feature index: the index also exports the
// settings component, which would otherwise end up in the initial bundle (#230).
import { ReadLaterApi } from '../../features/readlater/adapter/api/readLaterApi';
import { ReadLaterUseCases } from '../../features/readlater/usecase/readLaterUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    readlater: ReadLaterUseCases;
  }
}

export const service = defineService('readlater', (http) => new ReadLaterUseCases(new ReadLaterApi(http)));
