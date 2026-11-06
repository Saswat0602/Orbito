import { Tag } from '@/types';
import { storage, STORAGE_KEYS } from '../utils/storage';

const MOCK_TAGS: Tag[] = [
    { id: '1', name: 'design', color: 'purple', createdAt: new Date().toISOString() },
    { id: '2', name: 'ui', color: 'blue', createdAt: new Date().toISOString() },
    { id: '3', name: 'finance', color: 'green', createdAt: new Date().toISOString() },
    { id: '4', name: 'urgent', color: 'red', createdAt: new Date().toISOString() },
];

const initializeTags = (): Tag[] => {
    const stored = storage.get<Tag[]>(STORAGE_KEYS.TAGS);
    if (!stored || stored.length === 0) {
        storage.set(STORAGE_KEYS.TAGS, MOCK_TAGS);
        return MOCK_TAGS;
    }
    return stored;
};

export const getTags = async (): Promise<Tag[]> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    return initializeTags();
};

export const createTag = async (tagData: Omit<Tag, 'id' | 'createdAt'>): Promise<Tag> => {
    await new Promise(resolve => setTimeout(resolve, 300));

    const newTag: Tag = {
        ...tagData,
        id: Date.now().toString(),
        createdAt: new Date().toISOString(),
    };

    const tags = initializeTags();
    const updatedTags = [...tags, newTag];
    storage.set(STORAGE_KEYS.TAGS, updatedTags);

    return newTag;
};

export const deleteTag = async (id: string): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 300));

    const tags = initializeTags();
    const filteredTags = tags.filter(tag => tag.id !== id);
    storage.set(STORAGE_KEYS.TAGS, filteredTags);
};
