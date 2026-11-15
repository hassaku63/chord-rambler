import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { App } from './App';

function stubServer(answer: () => Promise<Response>) {
  vi.stubGlobal('fetch', vi.fn(answer));
}

describe('App', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('says it is checking until the server answers', () => {
    stubServer(() => new Promise<Response>(() => undefined));

    render(<App />);

    expect(screen.getByRole('status')).toHaveTextContent('サーバーを確認しています。');
  });

  it('shows the version once the server answers', async () => {
    stubServer(() =>
      Promise.resolve(new Response(JSON.stringify({ status: 'ok', version: 'abc1234' }))),
    );

    render(<App />);

    expect(await screen.findByText('サーバーにつながっています（版 abc1234）。')).toHaveRole(
      'status',
    );
    expect(fetch).toHaveBeenCalledWith('/api/v1/health', expect.anything());
  });

  it.each([
    { name: 'the request fails', answer: () => Promise.reject(new Error('network down')) },
    {
      name: 'the server answers 500',
      answer: () => Promise.resolve(new Response('', { status: 500 })),
    },
    {
      name: 'the body is not what the contract describes',
      answer: () => Promise.resolve(new Response(JSON.stringify({ status: 'down' }))),
    },
  ])('says the server is unreachable when $name', async ({ answer }) => {
    stubServer(answer);

    render(<App />);

    expect(await screen.findByText('サーバーにつながりません。')).toHaveRole('status');
  });
});
