import { fireEvent, render, screen } from "@testing-library/react";
import type { AnchorHTMLAttributes, ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { Movie } from "@/types";

vi.mock("next/link", () => ({
  default: ({ href, children, ...rest }: { href: string; children: ReactNode } & AnchorHTMLAttributes<HTMLAnchorElement>) => <a href={href} {...rest}>{children}</a>,
}));
import MovieCard from "./MovieCard";

const movie: Movie = {
  id: "m1", title: "Heat", year: 1995, duration_seconds: 10200, container: "mkv", video_codec: "h264", audio_codec: "aac",
  resolution: "1080p", synopsis: "", genres: [], rating: 8.25, file_size: 1e9, has_poster: true, has_backdrop: true,
  added_at: "2026-01-01 00:00:00.000Z", thumb_v: 4,
};

describe("MovieCard", () => {
  it("links to the movie page", () => {
    render(<MovieCard m={movie} />);
    expect(screen.getByRole("link")).toHaveAttribute("href", "/movie/?id=m1");
  });

  it("shows the poster, with the version for cache busting", () => {
    render(<MovieCard m={movie} />);
    expect(screen.getByAltText("Heat")).toHaveAttribute("src", "/api/movies/m1/poster?v=4");
  });

  it("shows the title and year on a card when there is no poster", () => {
    render(<MovieCard m={{ ...movie, has_poster: false }} />);
    expect(screen.queryByAltText("Heat")).toBeNull();
    expect(screen.getAllByText("Heat").length).toBeGreaterThan(0);
    expect(screen.getAllByText("1995").length).toBeGreaterThan(0);
  });

  it("leaves out an unknown year and an unknown rating", () => {
    render(<MovieCard m={{ ...movie, year: 0, rating: 0 }} />);
    expect(screen.queryByText("0")).toBeNull();
    expect(screen.queryByText("0.0")).toBeNull();
  });

  it("shows the rating to one decimal", () => {
    render(<MovieCard m={movie} />);
    expect(screen.getByText("8.3")).toBeInTheDocument();
  });

  it("draws watch progress, capped at the full width", () => {
    const { container, rerender } = render(<MovieCard m={movie} progress={0.4} />);
    expect((container.querySelector(".bg-accent[style]") as HTMLElement).style.width).toBe("40%");
    rerender(<MovieCard m={movie} progress={3} />);
    expect((container.querySelector(".bg-accent[style]") as HTMLElement).style.width).toBe("100%");
    rerender(<MovieCard m={movie} progress={0} />);
    expect(container.querySelector(".bg-accent[style]")).toBeNull();
  });

  it("marks a watched movie", () => {
    render(<MovieCard m={movie} watched />);
    expect(screen.getByTitle("Watched")).toBeInTheDocument();
  });

  it("has a my list button only when it can toggle", () => {
    const { rerender } = render(<MovieCard m={movie} />);
    expect(screen.queryByRole("button")).toBeNull();
    const toggle = vi.fn();
    rerender(<MovieCard m={movie} onToggleList={toggle} inList={false} />);
    fireEvent.click(screen.getByRole("button", { name: "Add to my list" }));
    expect(toggle).toHaveBeenCalledWith("m1");
    rerender(<MovieCard m={movie} onToggleList={toggle} inList />);
    expect(screen.getByRole("button", { name: "Remove from my list" })).toHaveAttribute("aria-pressed", "true");
  });

  it("uses a fixed width in a carousel only", () => {
    expect(render(<MovieCard m={movie} fixed />).container.firstElementChild!.className).toContain("w-36");
    expect(render(<MovieCard m={movie} />).container.firstElementChild!.className).toContain("w-full");
  });
});
