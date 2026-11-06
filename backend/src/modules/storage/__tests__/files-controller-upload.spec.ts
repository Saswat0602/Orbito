import { BadRequestException } from '@nestjs/common';

// storage.service.ts pulls in s3.service.ts, which imports `mime` v4 - ESM-only,
// unloadable by jest's CJS runtime. Nothing here touches S3, so stub it out.
jest.mock('../s3.service', () => ({ S3Service: class {} }));

import { FilesController } from '../files.controller';

// Guards the three rules that closed the "no size limit, no type filter,
// arbitrary folder" hole on POST /uploads/upload/:folder.
describe('FilesController.uploadFile', () => {
  const storageService = { saveFile: jest.fn(), isUsingS3: () => false } as any;
  const configService = {
    get: (_key: string, fallback: string) => fallback,
  } as any;

  const controller = new FilesController(storageService, configService);
  const image = (overrides: Partial<Express.Multer.File> = {}) =>
    ({ mimetype: 'image/png', size: 1024, ...overrides }) as Express.Multer.File;

  beforeEach(() => storageService.saveFile.mockReset());

  it('rejects any folder other than avatar', async () => {
    await expect(controller.uploadFile('tasks', image())).rejects.toBeInstanceOf(
      BadRequestException,
    );
    expect(storageService.saveFile).not.toHaveBeenCalled();
  });

  it('rejects a disallowed mime type', async () => {
    await expect(
      controller.uploadFile('avatar', image({ mimetype: 'text/html' })),
    ).rejects.toBeInstanceOf(BadRequestException);
  });

  it('rejects a file over the configured size limit', async () => {
    await expect(
      controller.uploadFile('avatar', image({ size: 50 * 1024 * 1024 + 1 })),
    ).rejects.toBeInstanceOf(BadRequestException);
  });

  it('rejects a missing file', async () => {
    await expect(controller.uploadFile('avatar', undefined as any)).rejects.toBeInstanceOf(
      BadRequestException,
    );
  });

  it('saves a valid avatar image', async () => {
    storageService.saveFile.mockResolvedValue({
      url: '/avatar/a.png',
      key: 'avatar/a.png',
      size: 1024,
    });
    const result = await controller.uploadFile('avatar', image());
    expect(storageService.saveFile).toHaveBeenCalledWith(expect.anything(), 'avatar');
    expect(result.key).toBe('avatar/a.png');
  });

  it('accepts a file at the 50MB boundary', async () => {
    storageService.saveFile.mockResolvedValue({ url: null, key: 'avatar/big.png', size: 0 });
    await expect(
      controller.uploadFile('avatar', image({ size: 50 * 1024 * 1024 })),
    ).resolves.toBeDefined();
  });
});

describe('FilesController construction', () => {
  const storageService = {} as any;

  it('throws on invalid MAX_AVATAR_SIZE', () => {
    const badConfig = { get: () => 'not-a-number' } as any;
    expect(() => new FilesController(storageService, badConfig)).toThrow('MAX_AVATAR_SIZE');
  });

  it('rejects zero and negative values the same way', () => {
    for (const bad of ['0', '-5']) {
      const badConfig = { get: () => bad } as any;
      expect(() => new FilesController(storageService, badConfig)).toThrow('MAX_AVATAR_SIZE');
    }
  });
});
