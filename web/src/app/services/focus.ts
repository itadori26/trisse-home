import { browserNotifier, FocusApi } from '../../features/focus/adapter/api/focusApi';
import { FocusUseCases } from '../../features/focus/usecase/focusUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    focus: FocusUseCases;
  }
}

export const service = defineService(
  'focus',
  (http) => new FocusUseCases(new FocusApi(http), browserNotifier),
);
