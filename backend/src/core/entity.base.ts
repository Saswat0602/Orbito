export abstract class BaseEntity<T> {
  protected readonly _id: string;
  protected readonly props: T;
  private readonly _createdAt: Date;
  private _updatedAt: Date;

  constructor(props: T, id?: string, createdAt?: Date, updatedAt?: Date) {
    this._id = id ?? crypto.randomUUID();
    this.props = props;
    this._createdAt = createdAt ?? new Date();
    this._updatedAt = updatedAt ?? new Date();
  }

  get id(): string {
    return this._id;
  }

  get createdAt(): Date {
    return this._createdAt;
  }

  get updatedAt(): Date {
    return this._updatedAt;
  }

  public equals(object?: BaseEntity<T>): boolean {
    if (object == null || object === undefined) {
      return false;
    }
    if (this === object) {
      return true;
    }
    return this._id === object._id;
  }
}
