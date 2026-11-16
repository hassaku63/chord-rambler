/**
 * A stand-in for the AudioContext, which jsdom does not have. It records each
 * oscillator: its wave, frequency, when it starts and stops, and the changes
 * of loudness of the gain it goes through.
 */

type Change = [kind: 'set' | 'ramp', value: number, time: number];

class FakeGain {
  gain = {
    value: 1,
    changes: [] as Change[],
    setValueAtTime(value: number, time: number) {
      this.changes.push(['set', value, time]);
    },
    linearRampToValueAtTime(value: number, time: number) {
      this.changes.push(['ramp', value, time]);
    },
  };
  connected: unknown = null;
  connect(node: unknown) {
    this.connected = node;
    return node;
  }
  disconnect() {
    this.connected = null;
  }
}

export class FakeOscillator {
  type = 'sine';
  frequency = { value: 0 };
  startedAt: number | null = null;
  stoppedAt: number | null = null;
  /** Stopped at once, by stop() with no time. */
  cut = false;
  through: FakeGain | null = null;
  connect(node: FakeGain) {
    this.through = node;
    return node;
  }
  disconnect() {
    this.through = null;
  }
  start(time: number) {
    this.startedAt = time;
  }
  stop(time?: number) {
    if (time === undefined) {
      this.cut = true;
    } else {
      this.stoppedAt = time;
    }
  }
  /** The changes of loudness of the voice's own gain. */
  get changes(): Change[] {
    return this.through?.gain.changes ?? [];
  }
}

export class FakeAudioContext {
  currentTime = 1;
  state: 'suspended' | 'running' | 'closed' = 'suspended';
  destination = {};
  oscillators: FakeOscillator[] = [];
  gains: FakeGain[] = [];
  resume() {
    this.state = 'running';
    return Promise.resolve();
  }
  close() {
    this.state = 'closed';
    return Promise.resolve();
  }
  createOscillator() {
    const oscillator = new FakeOscillator();
    this.oscillators.push(oscillator);
    return oscillator;
  }
  createGain() {
    const gain = new FakeGain();
    this.gains.push(gain);
    return gain;
  }
}

/** A factory for the player that hands out one fake, and the fake to look at. */
export function fakeAudio() {
  const context = new FakeAudioContext();
  let made = 0;
  const create = () => {
    made += 1;
    return context as unknown as AudioContext;
  };
  return { context, create, made: () => made };
}
