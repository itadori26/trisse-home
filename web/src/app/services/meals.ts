import { MealsApi } from '../../features/meals/adapter/api/mealsApi';
import { MealsUseCases } from '../../features/meals/usecase/mealsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    meals: MealsUseCases;
  }
}

export const service = defineService('meals', (http) => new MealsUseCases(new MealsApi(http)));
