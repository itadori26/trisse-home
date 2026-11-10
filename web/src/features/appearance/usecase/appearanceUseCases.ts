import type { Appearance } from '../domain/appearance';

/** Output port: appearance backend. */
export interface AppearanceGateway {
  load(): Promise<Appearance>;
  save(a: Appearance): Promise<Appearance>;
}

/** Application logic of the appearance feature. */
export class AppearanceUseCases {
  constructor(private readonly gateway: AppearanceGateway) {}

  load(): Promise<Appearance> {
    return this.gateway.load();
  }

  /** Saves trimmed preferences. */
  save(a: Appearance): Promise<Appearance> {
    return this.gateway.save({
      ...a,
      accentColor: a.accentColor.trim(),
      backgroundImage: a.backgroundImage.trim(),
    });
  }
}
