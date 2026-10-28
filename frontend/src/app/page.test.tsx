import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("Home", () => {
  it("renders the calculator heading and form", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", { name: "Calculator" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Calculate" }),
    ).toBeInTheDocument();
  });
});
