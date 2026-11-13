import { CountdownApi } from '../../features/countdown/adapter/api/countdownApi';
import { CountdownUseCases } from '../../features/countdown/usecase/countdownUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    countdown: CountdownUseCases;
  }
}

export const service = defineService('countdown', (http) => new CountdownUseCases(new CountdownApi(http)));
