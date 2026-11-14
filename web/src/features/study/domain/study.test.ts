import { describe, expect, it, vi } from 'vitest';
import { emptyModule, formatEcts, formatGrade, validateModule } from './study';
import { StudyUseCases, type StudyGateway } from '../usecase/studyUseCases';

describe('study domain (#57)', () => {
  it('validates the module form', () => {
    expect(validateModule({ ...emptyModule(), name: 'Analysis' })).toBeNull();
    expect(validateModule({ ...emptyModule(), name: ' ' })).not.toBeNull();
    expect(validateModule({ ...emptyModule(), name: 'x', ects: 2.25 })).not.toBeNull();
    expect(validateModule({ ...emptyModule(), name: 'x', ects: 0 })).not.toBeNull();
    expect(validateModule({ ...emptyModule(), name: 'x', grade: 2 })).not.toBeNull(); // open with grade
    expect(validateModule({ ...emptyModule(), name: 'x', status: 'passed', grade: 5 })).not.toBeNull();
    expect(validateModule({ ...emptyModule(), name: 'x', status: 'passed', grade: 1.3 })).toBeNull();
  });

  it('formats grades and credits in the locale', () => {
    expect(formatGrade(1.5, 'de')).toBe('1,5');
    expect(formatGrade(1.666666, 'en')).toBe('1.67');
    expect(formatGrade(2, 'en')).toBe('2.0');
    expect(formatEcts(7.5, 'de')).toBe('7,5');
    expect(formatEcts(180, 'en')).toBe('180');
  });

  it('drops the grade of modules that are not passed', async () => {
    const g = {
      modules: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
      progress: vi.fn(),
      totalEcts: vi.fn(),
      setTotalEcts: vi.fn(),
    } satisfies StudyGateway;
    const uc = new StudyUseCases(g);
    await uc.save({ name: ' Analysis ', ects: 9, semester: 1, status: 'in_progress', grade: 1.3 });
    expect(g.create).toHaveBeenCalledWith({
      name: 'Analysis',
      ects: 9,
      semester: 1,
      status: 'in_progress',
      grade: null,
    });
    await uc.save({ name: 'DB', ects: 6, semester: 2, status: 'passed', grade: 1.7 }, 'a');
    expect(g.update).toHaveBeenCalledWith('a', {
      name: 'DB',
      ects: 6,
      semester: 2,
      status: 'passed',
      grade: 1.7,
    });
    await expect(uc.save({ ...emptyModule(), name: '' })).rejects.toThrow();
  });
});
