import type { HttpClient } from '../../../../shared/api-client/http';
import type { ClearedDTO, ItemDTO, SettingsDTO } from '../../../../shared/api-client/generated/shopping';
import type {
  ItemStateDTO,
  NewItemDTO,
  SettingsInputDTO,
} from '../../../../shared/api-client/generated/shopping-requests';
import type { ShoppingItem, ShoppingSettings } from '../../domain/shopping';
import type { ShoppingGateway } from '../../usecase/shoppingUseCases';

const item = (id: string) => `/api/shopping/items/${encodeURIComponent(id)}`;

/** ShoppingGateway backed by the REST API. */
export class ShoppingApi implements ShoppingGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<ShoppingSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/shopping/settings')) };
  }
  async saveSettings(listUrl: string, username: string, password: string): Promise<ShoppingSettings> {
    const body: SettingsInputDTO = { listUrl, username, password };
    return { ...(await this.http.put<SettingsDTO>('/api/shopping/settings', body)) };
  }
  async items(): Promise<ShoppingItem[]> {
    return (await this.http.get<ItemDTO[]>('/api/shopping/items')).map((i) => ({ ...i }));
  }
  async add(summary: string): Promise<ShoppingItem> {
    const body: NewItemDTO = { summary };
    return { ...(await this.http.post<ItemDTO>('/api/shopping/items', body)) };
  }
  setDone(id: string, done: boolean): Promise<void> {
    const body: ItemStateDTO = { done };
    return this.http.put(item(id), body);
  }
  remove(id: string): Promise<void> {
    return this.http.delete(item(id));
  }
  async clearCompleted(): Promise<number> {
    return (await this.http.post<ClearedDTO>('/api/shopping/clear-completed')).deleted;
  }
}
