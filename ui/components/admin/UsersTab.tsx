"use client";
import { useState } from "react";
import { USERS_PAGE_SIZE } from "@/constants/admin";
import { useDebounced } from "@/hooks/useDebounced";
import { usePagedList } from "@/hooks/usePagedList";
import { useSentinel } from "@/hooks/useSentinel";
import { adminApi } from "@/lib/adminApi";
import type { AdminUser, Page } from "@/types";
import { fieldClass } from "./ui";
import UserEditor from "./UserEditor";

export default function UsersTab() {
  const [q, setQ] = useState("");
  const query = useDebounced(q.trim());
  const [editing, setEditing] = useState<AdminUser | null>(null);

  const { items, total, hasMore, loading, error, loadMore, reload } = usePagedList<AdminUser>(
    (offset) => adminApi<Page<AdminUser>>(`/api/admin/users?limit=${USERS_PAGE_SIZE}&offset=${offset}&q=${encodeURIComponent(query)}`),
    query,
  );
  const sentinel = useSentinel(loadMore, hasMore && !loading, items.length);

  return (
    <section className="mt-6">
      <div className="flex flex-wrap items-center gap-4">
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Search profile, name or email" aria-label="Search profiles"
          className={`${fieldClass} mt-0 max-w-xs`} />
        <p className="text-sm text-dim">Showing {items.length} of {total} profiles</p>
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="text-dim">
            <tr><th className="py-2">Profile</th><th>Name</th><th>Username</th><th>Devices</th><th>External</th><th>Status</th><th /></tr>
          </thead>
          <tbody>
            {items.map((u) => (
              <tr key={u.id} className="border-t border-edge">
                <td className="py-2 font-medium">{u.room_number || "-"}</td>
                <td>{u.display_name}</td>
                <td className="text-dim">{u.username}</td>
                <td className="tabular-nums">{u.online} / {u.max_connections}</td>
                <td className="text-dim">{u.allow_external ? "Allowed" : "Off"}</td>
                <td>
                  {!u.active ? <span className="rounded-full bg-red-500/20 px-2.5 py-0.5 text-xs text-red-300">Disabled</span>
                    : u.online > 0 ? <span className="rounded-full bg-emerald-500/20 px-2.5 py-0.5 text-xs text-emerald-300">Watching</span>
                    : <span className="rounded-full bg-white/10 px-2.5 py-0.5 text-xs text-dim">Idle</span>}
                </td>
                <td className="text-right"><button onClick={() => setEditing(u)} className="rounded border border-edge px-3 py-1 hover:border-accent">Edit</button></td>
              </tr>
            ))}
          </tbody>
        </table>
        {!loading && !error && items.length === 0 && <p className="mt-6 text-dim">{query ? "No profiles match that search." : "No profiles yet. Create some in the Profiles tab."}</p>}
      </div>

      {error && (
        <p role="alert" className="mt-3 text-sm text-red-400">
          {error} <button onClick={reload} className="underline">Try again</button>
        </p>
      )}
      {loading && <p className="mt-4 text-sm text-dim">Loading</p>}
      {hasMore && <div ref={sentinel} className="h-8" aria-hidden />}
      {!hasMore && items.length > USERS_PAGE_SIZE && <p className="mt-4 text-center text-xs text-dim">That is every profile.</p>}

      {editing && <UserEditor user={editing} onClose={() => setEditing(null)} onSaved={reload} />}
    </section>
  );
}
