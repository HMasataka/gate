import { describe, it, expect } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { ToastProvider, useToast } from "./Toast";

function TestTrigger() {
  const { showToast } = useToast();
  return (
    <button onClick={() => showToast("Test message", "success")}>
      Trigger
    </button>
  );
}

describe("Toast", () => {
  it("renders toast when showToast is called", async () => {
    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    );

    await act(async () => {
      screen.getByText("Trigger").click();
    });

    expect(screen.getByText("Test message")).toBeInTheDocument();
  });

  it("renders error toast with correct class", async () => {
    function ErrorTrigger() {
      const { showToast } = useToast();
      return (
        <button onClick={() => showToast("Error occurred", "error")}>
          Error
        </button>
      );
    }

    render(
      <ToastProvider>
        <ErrorTrigger />
      </ToastProvider>,
    );

    await act(async () => {
      screen.getByText("Error").click();
    });

    const toast = screen.getByText("Error occurred");
    expect(toast).toHaveClass("toast", "error");
  });
});
