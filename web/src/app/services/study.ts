// Imported from the files, not the feature index: the index also exports the
// settings component, which would otherwise end up in the initial bundle (#230).
import { StudyApi } from '../../features/study/adapter/api/studyApi';
import { StudyUseCases } from '../../features/study/usecase/studyUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    study: StudyUseCases;
  }
}

export const service = defineService('study', (http) => new StudyUseCases(new StudyApi(http)));
