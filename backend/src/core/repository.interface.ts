export interface IRepository<T> {
  findById(id: string): Promise<T | null>;
  findAll(filter?: Record<string, any>): Promise<T[]>;
  save(entity: T): Promise<T>;
  delete(id: string): Promise<boolean>;
}
