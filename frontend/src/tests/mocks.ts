import { vi } from "vitest";

const getMock = vi.fn();
const postMock = vi.fn();
const putMock = vi.fn();
const deleteMock = vi.fn();

export const api = {
  get: getMock,
  post: postMock,
  put: putMock,
  delete: deleteMock,
}
vi.mock("@/api", () => ({
  api: {
    get: (...args: unknown[]) => getMock(...args),
    post: (...args: unknown[]) => postMock(...args),
    put: (...args: unknown[]) => putMock(...args),
    delete: (...args: unknown[]) => deleteMock(...args),
  },
}));

const toastAddMock = vi.fn();

export const toast = {
  add: toastAddMock,
};
vi.mock("primevue/usetoast", () => ({
  useToast: () => ({ add: (...args: unknown[]) => toastAddMock(...args) }),
}));

const pushMock = vi.fn();
const currentRoute = { value: { params: { id: "plan-1" } } };

export const router = {
  push: pushMock,
  currentRoute,
};
vi.mock("vue-router", () => ({
  useRouter: () => ({
    push: (...args: unknown[]) => pushMock(...args),
    currentRoute,
  }),
}));
