import { describe, expect, it } from 'vitest';
import { parseHealth } from './client';

describe('parseHealth', () => {
  it('accepts the body the contract describes', () => {
    expect(parseHealth({ status: 'ok', version: 'abc1234' })).toEqual({
      status: 'ok',
      version: 'abc1234',
    });
  });

  it.each([
    { name: 'null', body: null },
    { name: 'a string', body: 'ok' },
    { name: 'no version', body: { status: 'ok' } },
    { name: 'no status', body: { version: 'abc1234' } },
    { name: 'another status', body: { status: 'down', version: 'abc1234' } },
    { name: 'a version that is not a string', body: { status: 'ok', version: 1 } },
  ])('rejects $name', ({ body }) => {
    expect(() => parseHealth(body)).toThrow('unexpected body');
  });
});
