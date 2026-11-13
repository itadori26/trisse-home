import { afterEach, describe, expect, it, vi } from 'vitest';
import { portal } from './portal';

/** Minimal stand-ins: tests run without a DOM (no extra dependency, ADR 0011). */
function fakeNode() {
  return {
    parent: 'from' as string | null,
    remove: vi.fn(function (this: { parent: string | null }) {
      this.parent = null;
    }),
  };
}

function stubDocument(ids: string[]) {
  vi.stubGlobal('document', {
    getElementById: (id: string) =>
      ids.includes(id) ? { appendChild: (n: { parent: string | null }) => (n.parent = id) } : null,
  });
}

afterEach(() => vi.unstubAllGlobals());

describe('portal', () => {
  it('moves the node into the target and removes it on destroy', () => {
    stubDocument(['to']);
    const node = fakeNode();
    const action = portal(node as unknown as HTMLElement, 'to');
    expect(node.parent).toBe('to');
    action.destroy();
    expect(node.remove).toHaveBeenCalled();
  });

  it('leaves the node in place without a target', () => {
    stubDocument([]);
    const node = fakeNode();
    portal(node as unknown as HTMLElement, 'missing');
    expect(node.parent).toBe('from');
  });
});
