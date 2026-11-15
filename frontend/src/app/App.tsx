import type { BackendHealth } from './useBackendHealth';
import { useBackendHealth } from './useBackendHealth';

function statusText(health: BackendHealth): string {
  switch (health.state) {
    case 'checking':
      return 'サーバーを確認しています。';
    case 'connected':
      return `サーバーにつながっています（版 ${health.version}）。`;
    case 'unreachable':
      return 'サーバーにつながりません。';
  }
}

/** The application shell. For now it only shows whether the API server answers. */
export function App() {
  const health = useBackendHealth();

  return (
    <main>
      <h1>chord-rambler</h1>
      <p role="status">{statusText(health)}</p>
    </main>
  );
}
