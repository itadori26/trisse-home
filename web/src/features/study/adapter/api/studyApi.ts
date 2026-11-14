import type { HttpClient } from '../../../../shared/api-client/http';
import type { ModuleDTO, ProgressDTO, SettingsDTO } from '../../../../shared/api-client/generated/study';
import type { ModuleInput, Progress, StudyModule } from '../../domain/study';
import type { StudyGateway } from '../../usecase/studyUseCases';

const toModule = (d: ModuleDTO): StudyModule => ({ ...d, grade: d.grade ?? null });

/** StudyGateway backed by the REST API. */
export class StudyApi implements StudyGateway {
  constructor(private readonly http: HttpClient) {}

  async modules(): Promise<StudyModule[]> {
    return (await this.http.get<ModuleDTO[]>('/api/study/modules')).map(toModule);
  }
  async create(m: ModuleInput): Promise<StudyModule> {
    return toModule(await this.http.post<ModuleDTO>('/api/study/modules', m));
  }
  async update(id: string, m: ModuleInput): Promise<StudyModule> {
    return toModule(await this.http.put<ModuleDTO>(`/api/study/modules/${encodeURIComponent(id)}`, m));
  }
  remove(id: string): Promise<void> {
    return this.http.delete(`/api/study/modules/${encodeURIComponent(id)}`);
  }
  async progress(): Promise<Progress> {
    const d = await this.http.get<ProgressDTO>('/api/study/progress');
    return { ...d, average: d.average ?? null, inProgress: d.inProgress.map(toModule) };
  }
  async totalEcts(): Promise<number> {
    return (await this.http.get<SettingsDTO>('/api/study/settings')).totalEcts;
  }
  async setTotalEcts(v: number): Promise<number> {
    return (await this.http.put<SettingsDTO>('/api/study/settings', { totalEcts: v })).totalEcts;
  }
}
