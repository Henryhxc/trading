import { render, screen } from "../test-utils";
import { describe, expect, it, vi } from "vitest";
import { AuthPage } from "./AuthPage";

vi.mock("../app/SessionProvider", () => ({
  useSession: () => ({ status: "anonymous", user: null, acceptUser: vi.fn(), clearUser: vi.fn(), retry: vi.fn() }),
}));

describe("AuthPage branding", () => {
  it("uses the light terminal theme with the supplied USMART logo", () => {
    render(<AuthPage mode="login" />);

    expect(screen.getByRole("main").classList.contains("terminal-auth")).toBe(true);
    expect(screen.getByRole("img", { name: "USMART 盈立證券" }).classList.contains("terminal-auth-logo")).toBe(true);
  });
});
