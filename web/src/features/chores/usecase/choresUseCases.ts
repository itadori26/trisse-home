import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import { parseAssignees, type Chore, type ChoreInput } from '../domain/chores';

/** The fields the backend stores; assignees as a list. */
export interface ChoreData {
  title: string;
  everyDays: number;
  due: string;
  assignees: string[];
}

/** Output port: chore plan backend. */
export interface ChoresGateway {
  list(): Promise<Chore[]>;
  create(data: ChoreData, visibility?: Visibility): Promise<Chore>;
  update(id: string, data: ChoreData, visibility?: Visibility): Promise<Chore>;
  remove(id: string): Promise<void>;
  done(id: string): Promise<Chore>;
  shareGroups(): Promise<ShareGroup[]>;
}

const toData = (i: ChoreInput): ChoreData => ({
  title: i.title.trim(),
  everyDays: i.everyDays,
  due: i.due,
  assignees: parseAssignees(i.assignees),
});

/** Application logic of the chore plan (#683). */
export class ChoresUseCases {
  constructor(private readonly gateway: ChoresGateway) {}

  list(): Promise<Chore[]> {
    return this.gateway.list();
  }

  /** Creates (no id) or changes a chore. */
  save(id: string | null, input: ChoreInput, visibility?: Visibility): Promise<Chore> {
    const data = toData(input);
    return id ? this.gateway.update(id, data, visibility) : this.gateway.create(data, visibility);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  /** Ticks a chore off; it moves on and passes to the next person. */
  done(id: string): Promise<Chore> {
    return this.gateway.done(id);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }
}
