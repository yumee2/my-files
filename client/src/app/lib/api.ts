export interface StoredFile {
  id: string;
  name: string;
  size: number;
  type: string;
  uploadedAt?: number;
}

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request(path: string, init: RequestInit = {}) {
  const response = await fetch(path, {
    ...init,
    credentials: 'include',
  });

  if (!response.ok) {
    const message = await response.text();
    throw new ApiError(response.status, message || response.statusText);
  }

  return response;
}

export async function login(password: string) {
  const formData = new FormData();
  formData.set('password', password);

  await request('/login', {
    method: 'POST',
    body: formData,
  });
}

export async function getFiles(): Promise<StoredFile[]> {
  const response = await request('/files');
  return response.json();
}

export async function uploadFile(file: File): Promise<string> {
  const formData = new FormData();
  formData.set('file', file);

  const response = await request('/files', {
    method: 'POST',
    body: formData,
  });
  const data = await response.json();
  return data.id;
}

export async function deleteFile(id: string) {
  await request(`/files/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

export async function downloadFile(file: StoredFile) {
  const response = await request(`/files/${encodeURIComponent(file.id)}`);
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');

  link.href = url;
  link.download = file.name;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
