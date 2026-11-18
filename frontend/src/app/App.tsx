import { ProgressionScreen } from '@/features/progression/ProgressionScreen';
import type { BackendHealth } from './useBackendHealth';
import { useBackendHealth } from './useBackendHealth';
import './app.css';

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

/** The application shell: whether the API server answers, and the screen of the progression. */
export function App() {
  const health = useBackendHealth();

  return (
    <main>
      <header className="app-header">
        <h1>chord-rambler</h1>
        <p role="status">{statusText(health)}</p>
      </header>
      <ProgressionScreen />
    </main>
  );
}
