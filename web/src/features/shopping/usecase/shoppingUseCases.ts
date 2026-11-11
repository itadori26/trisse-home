import type { ShoppingItem, ShoppingSettings } from '../domain/shopping';

/** Output port: shopping list backend. */
export interface ShoppingGateway {
  settings(): Promise<ShoppingSettings>;
  saveSettings(listUrl: string, username: string, password: string): Promise<ShoppingSettings>;
  items(): Promise<ShoppingItem[]>;
  add(summary: string): Promise<ShoppingItem>;
  setDone(id: string, done: boolean): Promise<void>;
  remove(id: string): Promise<void>;
  clearCompleted(): Promise<number>;
}

/** Application logic of the shopping list feature. */
export class ShoppingUseCases {
  constructor(private readonly gateway: ShoppingGateway) {}

  settings(): Promise<ShoppingSettings> {
    return this.gateway.settings();
  }
  saveSettings(listUrl: string, username: string, password: string): Promise<ShoppingSettings> {
    return this.gateway.saveSettings(listUrl.trim(), username.trim(), password);
  }
  items(): Promise<ShoppingItem[]> {
    return this.gateway.items();
  }
  /** Adds an item; empty input is ignored. */
  async add(summary: string): Promise<ShoppingItem | null> {
    const s = summary.trim();
    return s ? this.gateway.add(s) : null;
  }
  setDone(id: string, done: boolean): Promise<void> {
    return this.gateway.setDone(id, done);
  }
  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
  clearCompleted(): Promise<number> {
    return this.gateway.clearCompleted();
  }
}
