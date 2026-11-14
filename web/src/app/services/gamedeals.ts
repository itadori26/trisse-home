import { GamedealsApi } from '../../features/gamedeals/adapter/api/gamedealsApi';
import { GamedealsUseCases } from '../../features/gamedeals/usecase/gamedealsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    gamedeals: GamedealsUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('gamedeals', (http) => new GamedealsUseCases(new GamedealsApi(http)));
