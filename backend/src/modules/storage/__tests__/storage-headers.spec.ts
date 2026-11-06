import { Response } from 'express';

// S3Service pulls in `mime`, which is ESM-only and cannot be required by jest's
// CJS runtime. Nothing here touches S3, so stub the whole module out.
jest.mock('../s3.service', () => ({ S3Service: class {} }));

import { StorageService } from '../storage.service';

// Guards the one rule that keeps an upload from becoming stored XSS: a streamed
// file always carries an explicit content type, and only types that cannot run
// script are allowed to render inline.
describe('StorageService content headers', () => {
  const service = new StorageService({ get: () => './uploads' } as any, {} as any);

  const call = (key: string) => {
    const headers: Record<string, string> = {};
    const res = { setHeader: (k: string, v: string) => (headers[k] = v) } as unknown as Response;
    (service as any).setContentHeaders(key, res);
    return headers;
  };

  it('never leaves the type for the browser to sniff', () => {
    expect(call('tasks/1/a.png')['X-Content-Type-Options']).toBe('nosniff');
  });

  it('forces script-capable uploads to download instead of rendering', () => {
    for (const name of ['x.svg', 'x.html', 'x.htm', 'x.xhtml', 'x.js', 'x.xml']) {
      expect(call(`tasks/1/${name}`)['Content-Disposition']).toMatch(/^attachment;/);
    }
  });

  it('still renders real images and pdfs inline', () => {
    for (const name of ['a.png', 'a.jpg', 'a.webp', 'a.pdf']) {
      expect(call(`tasks/1/${name}`)['Content-Disposition']).toMatch(/^inline;/);
    }
  });

  it('falls back to octet-stream for unknown extensions', () => {
    const headers = call('tasks/1/a.weirdext');
    expect(headers['Content-Type']).toBe('application/octet-stream');
    expect(headers['Content-Disposition']).toMatch(/^attachment;/);
  });

  it('strips quotes and newlines that would break out of the header', () => {
    expect(call('tasks/1/a".png')['Content-Disposition']).toBe('inline; filename="a.png"');
  });
});
