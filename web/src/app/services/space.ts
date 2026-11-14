// Imported from the files, not the feature index: the index also exports the
// settings component, which would otherwise end up in the initial bundle (#230).
import { SpaceApi } from '../../features/space/adapter/api/spaceApi';
import { SpaceUseCases } from '../../features/space/usecase/spaceUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    space: SpaceUseCases;
  }
}

export const service = defineService('space', (http) => new SpaceUseCases(new SpaceApi(http)));
