import { GamedealsApi, GamedealsUseCases } from '../../features/gamedeals';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    gamedeals: GamedealsUseCases;
  }
}

export const service = defineService('gamedeals', (http) => new GamedealsUseCases(new GamedealsApi(http)));
