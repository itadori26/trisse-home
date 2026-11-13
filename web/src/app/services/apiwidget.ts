// Imported from the files, not the feature index: the index also exports the
// settings component, which would otherwise end up in the initial bundle.
import { ApiWidgetApi } from '../../features/apiwidget/adapter/api/apiWidgetApi';
import { ApiWidgetUseCases } from '../../features/apiwidget/usecase/apiWidgetUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    apiwidget: ApiWidgetUseCases;
  }
}

export const service = defineService('apiwidget', (http) => new ApiWidgetUseCases(new ApiWidgetApi(http)));
