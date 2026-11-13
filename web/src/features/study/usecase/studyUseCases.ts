import { type ModuleInput, type Progress, type StudyModule, validateModule } from '../domain/study';

/** Output port: study backend. */
export interface StudyGateway {
  modules(): Promise<StudyModule[]>;
  create(m: ModuleInput): Promise<StudyModule>;
  update(id: string, m: ModuleInput): Promise<StudyModule>;
  remove(id: string): Promise<void>;
  progress(): Promise<Progress>;
  totalEcts(): Promise<number>;
  setTotalEcts(v: number): Promise<number>;
}

/** Application logic of the study progress feature (#57). */
export class StudyUseCases {
  constructor(private readonly gateway: StudyGateway) {}

  modules(): Promise<StudyModule[]> {
    return this.gateway.modules();
  }

  /** Validates and saves; a grade is dropped unless the module is passed. */
  save(m: ModuleInput, id?: string): Promise<StudyModule> {
    const clean: ModuleInput = { ...m, name: m.name.trim(), grade: m.status === 'passed' ? m.grade : null };
    const problem = validateModule(clean);
    if (problem) return Promise.reject(new Error(problem));
    return id ? this.gateway.update(id, clean) : this.gateway.create(clean);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  progress(): Promise<Progress> {
    return this.gateway.progress();
  }

  totalEcts(): Promise<number> {
    return this.gateway.totalEcts();
  }

  setTotalEcts(v: number): Promise<number> {
    return this.gateway.setTotalEcts(v);
  }
}
