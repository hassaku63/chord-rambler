import { useEffect, useState } from 'react';
import { getHealth } from '@/shared/api/client';

/** Whether the API server answers. A version exists only once it has answered. */
export type BackendHealth =
  { state: 'checking' } | { state: 'connected'; version: string } | { state: 'unreachable' };

/** Asks the API server for its health once, when the component mounts. */
export function useBackendHealth(): BackendHealth {
  const [health, setHealth] = useState<BackendHealth>({ state: 'checking' });

  useEffect(() => {
    const controller = new AbortController();
    getHealth(controller.signal)
      .then(({ version }) => setHealth({ state: 'connected', version }))
      .catch(() => {
        if (!controller.signal.aborted) {
          setHealth({ state: 'unreachable' });
        }
      });
    return () => controller.abort();
  }, []);

  return health;
}
