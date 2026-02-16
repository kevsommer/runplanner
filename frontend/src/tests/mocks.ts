import { vi } from "vitest";

const postMock = vi.fn();
const putMock = vi.fn();
const deleteMock = vi.fn();

export const api = {
  post: postMock,
  put: putMock,
  delete: deleteMock,
}
vi.mock("@/api", () => ({
  api: {
    post: (...args: unknown[]) => postMock(...args),
    put: (...args: unknown[]) => putMock(...args),
    delete: (...args: unknown[]) => deleteMock(...args),
  },
}));

const pushMock = vi.fn();

export const router = {
  push: pushMock,
};
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: (...args: unknown[]) => pushMock(...args) }),
}));
