import { HabitsApi } from '../../features/habits/adapter/api/habitsApi';
import { HabitsUseCases } from '../../features/habits/usecase/habitsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    habits: HabitsUseCases;
  }
}

export const service = defineService('habits', (http) => new HabitsUseCases(new HabitsApi(http)));
