"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { itemsApi, type Comment, type Item, type ItemListResponse } from "@/services/api";

type FormState = {
  title: string;
  description: string;
};

const initialForm: FormState = { title: "", description: "" };

export default function ItemsPage() {
  const [items, setItems] = useState<Item[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [sortBy, setSortBy] = useState<"updatedAt" | "createdAt" | "title">(
    "updatedAt",
  );
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [meta, setMeta] = useState<Omit<ItemListResponse, "items">>({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPages: 1,
  });
  const [form, setForm] = useState<FormState>(initialForm);
  const [editingID, setEditingID] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [commentInputs, setCommentInputs] = useState<Record<string, string>>({});
  const [commentsByItem, setCommentsByItem] = useState<Record<string, Comment[]>>({});
  const [uploadingByItem, setUploadingByItem] = useState<Record<string, boolean>>({});

  const isEditing = useMemo(() => Boolean(editingID), [editingID]);

  const loadItems = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await itemsApi.list({ page, pageSize, search, sortBy, sortOrder });
      setItems(data.items);
      setMeta({
        page: data.page,
        pageSize: data.pageSize,
        total: data.total,
        totalPages: data.totalPages,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load items");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, search, sortBy, sortOrder]);

  useEffect(() => {
    void loadItems();
  }, [loadItems]);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      if (isEditing) {
        await itemsApi.update(editingID, form);
      } else {
        await itemsApi.create(form);
      }
      setForm(initialForm);
      setEditingID("");
      if (page !== 1) {
        setPage(1);
      } else {
        await loadItems();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save item");
    } finally {
      setSaving(false);
    }
  }

  async function onDelete(id: string) {
    setError("");
    try {
      await itemsApi.remove(id);
      await loadItems();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete item");
    }
  }

  async function loadComments(itemID: string) {
    try {
      const comments = await itemsApi.comments(itemID);
      setCommentsByItem((prev) => ({ ...prev, [itemID]: comments }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load comments");
    }
  }

  async function addComment(itemID: string) {
    const content = (commentInputs[itemID] ?? "").trim();
    if (!content) {
      return;
    }
    try {
      await itemsApi.addComment(itemID, { content });
      setCommentInputs((prev) => ({ ...prev, [itemID]: "" }));
      await loadComments(itemID);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add comment");
    }
  }

  return (
    <main className="space-y-4">
      <section className="stratyx-panel p-5">
        <h1 className="text-2xl font-semibold">Items Management</h1>
        <p className="mt-1 text-sm opacity-75">
          Full CRUD with persistent MongoDB storage.
        </p>

        <form onSubmit={onSubmit} className="mt-4 grid gap-3 md:max-w-2xl">
          <input
            required
            minLength={2}
            value={form.title}
            onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))}
            placeholder="Title"
            className="rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <textarea
            value={form.description}
            onChange={(event) =>
              setForm((prev) => ({ ...prev, description: event.target.value }))
            }
            placeholder="Description"
            className="min-h-24 rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <div className="flex gap-2">
            <button
              disabled={saving}
              className="rounded-md bg-[var(--primary)] px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
            >
              {saving ? "Saving..." : isEditing ? "Update Item" : "Create Item"}
            </button>
            {isEditing && (
              <button
                type="button"
                onClick={() => {
                  setEditingID("");
                  setForm(initialForm);
                }}
                className="rounded-md border border-[var(--border)] px-4 py-2 text-sm"
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Your Items</h2>
        <div className="mt-3 grid gap-3 md:grid-cols-4">
          <input
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            placeholder="Search title/description"
            className="rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <select
            value={sortBy}
            onChange={(event) =>
              setSortBy(event.target.value as "updatedAt" | "createdAt" | "title")
            }
            className="rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          >
            <option value="updatedAt">Sort by updated</option>
            <option value="createdAt">Sort by created</option>
            <option value="title">Sort by title</option>
          </select>
          <select
            value={sortOrder}
            onChange={(event) => setSortOrder(event.target.value as "asc" | "desc")}
            className="rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          >
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
          <div className="flex gap-2">
            <button
              onClick={() => {
                setPage(1);
                setSearch(searchInput.trim());
              }}
              className="rounded-md bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white"
            >
              Apply
            </button>
            <button
              onClick={() => {
                setSearchInput("");
                setSearch("");
                setPage(1);
              }}
              className="rounded-md border border-[var(--border)] px-3 py-2 text-sm"
            >
              Reset
            </button>
          </div>
        </div>
        {error && <p className="mt-3 text-sm text-red-500">{error}</p>}
        {loading ? (
          <p className="mt-3 text-sm opacity-70">Loading...</p>
        ) : items.length === 0 ? (
          <p className="mt-3 text-sm opacity-70">No items yet. Create your first item.</p>
        ) : (
          <div className="mt-3 space-y-2">
            {items.map((item) => (
              <article
                key={item.id}
                className="rounded-md border border-[var(--border)] bg-[var(--surface-muted)] p-3"
              >
                <h3 className="font-semibold">{item.title}</h3>
                <p className="mt-1 text-sm opacity-80">{item.description}</p>
                <p className="mt-2 text-xs opacity-60">
                  Updated: {new Date(item.updatedAt).toLocaleString()}
                </p>
                <div className="mt-3 flex gap-2">
                  <button
                    onClick={() => {
                      setEditingID(item.id);
                      setForm({ title: item.title, description: item.description });
                    }}
                    className="rounded-md border border-[var(--border)] px-3 py-1 text-xs"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => void onDelete(item.id)}
                    className="rounded-md border border-red-500 px-3 py-1 text-xs text-red-500"
                  >
                    Delete
                  </button>
                </div>
                <div className="mt-4 rounded-md bg-white/60 p-3">
                  <div className="mb-2">
                    <input
                      type="file"
                      onChange={async (event) => {
                        const file = event.target.files?.[0];
                        if (!file) {
                          return;
                        }
                        setUploadingByItem((prev) => ({ ...prev, [item.id]: true }));
                        try {
                          await itemsApi.uploadAttachment(item.id, file);
                        } catch (err) {
                          setError(err instanceof Error ? err.message : "Failed to upload file");
                        } finally {
                          setUploadingByItem((prev) => ({ ...prev, [item.id]: false }));
                        }
                      }}
                      className="text-xs"
                    />
                    {uploadingByItem[item.id] && <p className="text-xs opacity-70">Uploading...</p>}
                  </div>
                  <div className="mb-2 flex gap-2">
                    <input
                      value={commentInputs[item.id] ?? ""}
                      onChange={(event) =>
                        setCommentInputs((prev) => ({ ...prev, [item.id]: event.target.value }))
                      }
                      placeholder="Add comment (supports @email mentions)"
                      className="w-full rounded-md border border-[var(--border)] bg-transparent px-2 py-1 text-xs"
                    />
                    <button
                      onClick={() => void addComment(item.id)}
                      className="rounded-md bg-[var(--primary)] px-2 py-1 text-xs font-semibold text-white"
                    >
                      Send
                    </button>
                  </div>
                  <button
                    onClick={() => void loadComments(item.id)}
                    className="mb-2 text-xs text-[var(--primary)]"
                  >
                    Load comments
                  </button>
                  <div className="space-y-1">
                    {(commentsByItem[item.id] ?? []).map((comment) => (
                      <div key={comment.id} className="flex items-center justify-between gap-2">
                        <p className="text-xs opacity-80">{comment.content}</p>
                        <button
                          onClick={async () => {
                            const reason = prompt("Reason for report?") ?? "";
                            if (!reason.trim()) {
                              return;
                            }
                            try {
                              await itemsApi.reportComment(comment.id, reason);
                            } catch (err) {
                              setError(err instanceof Error ? err.message : "Failed to report comment");
                            }
                          }}
                          className="text-[10px] text-red-500"
                        >
                          Report
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}
        <div className="mt-4 flex items-center justify-between">
          <p className="text-xs opacity-70">
            Page {meta.page} / {meta.totalPages} - {meta.total} total items
          </p>
          <div className="flex items-center gap-2">
            <select
              value={pageSize}
              onChange={(event) => {
                setPageSize(Number(event.target.value));
                setPage(1);
              }}
              className="rounded-md border border-[var(--border)] bg-transparent px-2 py-1 text-xs"
            >
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={20}>20</option>
            </select>
            <button
              disabled={meta.page <= 1}
              onClick={() => setPage((prev) => Math.max(1, prev - 1))}
              className="rounded-md border border-[var(--border)] px-2 py-1 text-xs disabled:opacity-50"
            >
              Prev
            </button>
            <button
              disabled={meta.page >= meta.totalPages}
              onClick={() => setPage((prev) => Math.min(meta.totalPages, prev + 1))}
              className="rounded-md border border-[var(--border)] px-2 py-1 text-xs disabled:opacity-50"
            >
              Next
            </button>
          </div>
        </div>
      </section>
    </main>
  );
}
