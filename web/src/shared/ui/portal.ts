/**
 * Svelte action that renders an element into another place of the page,
 * e.g. the dashboard's tab bar into the top bar of the shell (#167). The
 * target is looked up by id; without it the element stays where it is.
 */
export function portal(node: HTMLElement, targetId: string) {
  const move = (id: string) => document.getElementById(id)?.appendChild(node);
  move(targetId);
  return {
    update: move,
    destroy: () => node.remove(),
  };
}
