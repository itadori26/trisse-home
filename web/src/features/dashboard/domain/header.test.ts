import { describe, expect, it } from 'vitest';
import type { Tab, Widget } from './dashboard';
import { firstWidgetConfig, showTabBar } from './header';

const widget = (id: string, kind: string, x: number, y: number): Widget => ({
  id,
  tabId: '',
  kind,
  x,
  y,
  w: 2,
  h: 2,
  config: { id },
});

const tab = (id: string, widgets: Widget[]): Tab => ({
  id,
  name: id,
  position: 0,
  background: '',
  rules: [],
  widgets,
  mobile: [],
  mobileCustom: false,
});

describe('showTabBar', () => {
  it('hides a single tab unless editing', () => {
    expect(showTabBar(1, false)).toBe(false);
    expect(showTabBar(1, true)).toBe(true);
    expect(showTabBar(2, false)).toBe(true);
    expect(showTabBar(0, false)).toBe(false);
  });
});

describe('firstWidgetConfig', () => {
  const tabs = [
    tab('home', [widget('w1', 'clock', 0, 0), widget('w2', 'weather', 4, 2), widget('w3', 'weather', 0, 2)]),
    tab('server', [widget('w4', 'weather', 0, 0)]),
  ];
  it('prefers the active tab and the top-left widget', () => {
    expect(firstWidgetConfig(tabs, 'home', 'weather')).toEqual({ id: 'w3' });
    expect(firstWidgetConfig(tabs, 'server', 'weather')).toEqual({ id: 'w4' });
  });
  it('falls back to other tabs and returns null when missing', () => {
    expect(firstWidgetConfig([...tabs].reverse(), 'nope', 'clock')).toEqual({ id: 'w1' });
    expect(firstWidgetConfig(tabs, 'home', 'radar')).toBeNull();
  });
});
