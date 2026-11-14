import type { Container } from './container';

/**
 * The compact, allow-listed context of an AI question (#162): status lines
 * from the same read-only use cases as the smart search, so the model sees
 * only what the user could see anyway – never settings or credentials.
 */
export async function buildAiContext(c: Container, modules: Set<string>, now = new Date()): Promise<string> {
  const parts: Promise<string[]>[] = [];
  const time = (d: Date) => d.toISOString().slice(0, 16).replace('T', ' ');
  if (modules.has('system')) {
    parts.push(
      c.system
        .snapshot()
        .then((s) => [
          `host: cpu ${Math.round(s.cpuPercent)} %, memory ${Math.round(s.memory.percent)} %, load ${s.load[0].toFixed(2)}`,
        ]),
    );
  }
  if (modules.has('services')) {
    parts.push(
      c.services
        .list()
        .then((list) => list.map((s) => `service ${s.name}: ${s.state}${s.error ? ` (${s.error})` : ''}`)),
    );
  }
  if (modules.has('docker')) {
    parts.push(
      c.docker
        .containers()
        .then((l) =>
          l.containers.map(
            (x) => `container ${x.name}: ${x.state}${x.health ? `, ${x.health}` : ''} (${x.status})`,
          ),
        ),
    );
  }
  if (modules.has('timeline')) {
    parts.push(
      c.timeline
        .list({ source: '', severity: '', limit: 40 })
        .then((events) =>
          events.map((e) => `${time(e.at)} ${e.severity} ${e.source} ${e.subject}: ${e.message}`),
        ),
    );
  }
  const lines = (await Promise.all(parts.map((p) => p.catch(() => [] as string[])))).flat();
  return [`now: ${time(now)} UTC`, ...lines].join('\n').slice(0, 6000);
}
