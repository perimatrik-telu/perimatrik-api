"""Run the same scenario against Laravel and Go, compare normalized responses."""
import json, re, sys, time, requests

PNG = bytes.fromhex(
    "89504e470d0a1a0a0000000d4948445200000001000000010806000000"
    "1f15c4890000000d49444154789c6360000002000154a24f5d0000000049454e44ae426082")
HTML = b"<html><script>alert(1)</script></html>"

ADMIN, OFFICER, BUDI, LAIN = (f"a0000000-0000-4000-8000-00000000000{i}" for i in range(1, 5))
ART1, ART_DRAFT, ART_FUTURE, ART_BUDI = (f"d0000000-0000-4000-8000-00000000000{i}" for i in range(1, 5))
NEWS1, NEWS2 = "e0000000-0000-4000-8000-000000000001", "e0000000-0000-4000-8000-000000000002"
GAL1 = "f0000000-0000-4000-8000-000000000001"
ANG1, ANG2 = "b0000000-0000-4000-8000-000000000001", "b0000000-0000-4000-8000-000000000002"
MEM1, MEM2 = "c0000000-0000-4000-8000-000000000001", "c0000000-0000-4000-8000-000000000002"
C1 = "70000000-0000-4000-8000-000000000001"
MISSING = "00000000-0000-4000-8000-0000000000ff"


def scenario(base):
    results, tok = [], {}

    def call(name, method, path, token=None, json_body=None, data=None, files=None, raw=False):
        h = {"Accept": "application/json"}
        if token:
            h["Authorization"] = f"Bearer {tok[token]}"
        r = requests.request(method, base + path, headers=h, json=json_body, data=data, files=files, allow_redirects=False)
        if raw:
            entry = {"status": r.status_code, "ctype": r.headers.get("Content-Type", "").split(";")[0],
                     "body": r.content}
        else:
            try:
                body = r.json()
            except Exception:
                body = {"<non-json>": r.text[:300]}
            entry = {"status": r.status_code, "body": body}
        results.append((name, entry))
        return entry

    def data_of(e):
        return e["body"].get("data") if isinstance(e["body"], dict) else None

    # ---------- public reads ----------
    call("root", "GET", "/")
    call("angkatan.list", "GET", "/api/v1/content/angkatan")
    call("article.list.guest", "GET", "/api/v1/content/article")
    call("article.list.page", "GET", "/api/v1/content/article?page=1&limit=1&filter=ASC")
    call("article.show.slug", "GET", "/api/v1/content/article/artikel-pertama-1700000000")
    call("article.show.draft.guest", "GET", "/api/v1/content/article/draft-budi-1700000001")
    call("article.show.future.guest", "GET", f"/api/v1/content/article/id/{ART_FUTURE}")
    call("article.show.id", "GET", f"/api/v1/content/article/id/{ART_BUDI}")
    call("article.show.missing", "GET", "/api/v1/content/article/nope")
    call("article.related", "GET", "/api/v1/content/article/artikel-pertama-1700000000/related")
    call("article.list.userid.guest", "GET", f"/api/v1/content/article?user_id={BUDI}")
    call("news.list", "GET", "/api/v1/content/news")
    call("news.list.user", "GET", f"/api/v1/content/news?user_id={OFFICER}")
    call("news.show", "GET", "/api/v1/content/news/berita-satu-1700000000")
    call("news.show.id", "GET", f"/api/v1/content/news/id/{NEWS2}")
    call("news.show.missing", "GET", f"/api/v1/content/news/id/{MISSING}")
    call("news.related", "GET", "/api/v1/content/news/berita-satu-1700000000/related")
    call("gallery.list", "GET", "/api/v1/content/gallery")
    call("gallery.list.user", "GET", f"/api/v1/content/gallery?user_id={BUDI}&limit=5")
    call("gallery.show", "GET", f"/api/v1/content/gallery/{GAL1}")
    call("gallery.show.missing", "GET", f"/api/v1/content/gallery/{MISSING}")
    call("image.show", "GET", "/api/v1/images/galleries/e2e-seed.png", raw=True)
    call("image.missing", "GET", "/api/v1/images/galleries/none.png")
    call("member.list", "GET", "/api/v1/content/member")
    call("member.show", "GET", f"/api/v1/content/member/{MEM1}")
    call("member.show.missing", "GET", f"/api/v1/content/member/{MISSING}")
    call("angkatan.show", "GET", f"/api/v1/content/angkatan/{ANG1}")
    call("angkatan.show.missing", "GET", f"/api/v1/content/angkatan/{MISSING}")
    call("comments.list", "GET", f"/api/v1/content/comments?article_id={ART1}")
    call("comments.list.nocontext", "GET", "/api/v1/content/comments")
    call("comments.list.invalid", "GET", "/api/v1/content/comments?article_id=xyz")
    call("settings.contact", "GET", "/api/v1/system/settings/contact")
    call("export.csv", "GET", "/api/v1/content/export-members?format=csv", raw=True)
    call("export.xlsx", "GET", "/api/v1/content/export-members", raw=True)
    call("share.article", "GET", "/share/article/artikel-pertama-1700000000", raw=True)
    call("share.news.missing", "GET", "/share/news/nope", raw=True)
    call("notfound", "GET", "/api/v1/nothing-here")

    # ---------- auth ----------
    call("auth.me.noauth", "GET", "/api/v1/auth/current-user")
    call("login.empty", "POST", "/api/v1/auth/login", json_body={})
    call("login.wrong", "POST", "/api/v1/auth/login", json_body={"email": "admin@e2e.test", "password": "salah"})
    for who, email in [("admin", "admin@e2e.test"), ("officer", "officer@e2e.test"), ("budi", "budi@e2e.test"), ("lain", "lain@e2e.test")]:
        e = call(f"login.{who}", "POST", "/api/v1/auth/login", json_body={"email": email, "password": "rahasia123"})
        tok[who] = data_of(e)["access_token"]
    call("auth.me", "GET", "/api/v1/auth/current-user", "budi")
    call("auth.user", "GET", "/api/v1/user", "admin")
    call("admin-only.member", "GET", "/api/v1/admin-only", "budi")
    call("admin-only.admin", "GET", "/api/v1/admin-only", "admin")
    call("verify.wrong", "POST", "/api/v1/auth/verify-password", "budi", json_body={"old_password": "x"})
    call("verify.ok", "POST", "/api/v1/auth/verify-password", "budi", json_body={"old_password": "rahasia123"})
    call("verify.invalid", "POST", "/api/v1/auth/verify-password", "budi", json_body={})
    call("changepw.invalid", "POST", "/api/v1/auth/change-password", "budi",
         json_body={"old_password": "rahasia123", "password": "short", "password_confirmation": "x"})
    call("changepw.wrongold", "POST", "/api/v1/auth/change-password", "lain",
         json_body={"old_password": "nope", "password": "barubaru123", "password_confirmation": "barubaru123"})
    call("changepw.ok", "POST", "/api/v1/auth/change-password", "lain",
         json_body={"old_password": "rahasia123", "password": "barubaru123", "password_confirmation": "barubaru123"})
    call("login.lain.newpw", "POST", "/api/v1/auth/login", json_body={"email": "lain@e2e.test", "password": "barubaru123"})

    # ---------- register ----------
    call("register.invalid", "POST", "/api/v1/auth/register", json_body={"email": "bad"})
    call("register.unknown", "POST", "/api/v1/auth/register", json_body={
        "no_anggota": "999", "nama_anggota": "X", "email": "x@e2e.test", "password": "rahasia123", "password_confirmation": "rahasia123"})
    call("register.namemismatch", "POST", "/api/v1/auth/register", json_body={
        "no_anggota": "002", "nama_anggota": "Bukan Siti", "email": "x@e2e.test", "password": "rahasia123", "password_confirmation": "rahasia123"})
    call("register.linked", "POST", "/api/v1/auth/register", json_body={
        "no_anggota": "001", "nama_anggota": "budi santoso", "email": "x@e2e.test", "password": "rahasia123", "password_confirmation": "rahasia123"})
    call("register.ok", "POST", "/api/v1/auth/register", json_body={
        "no_anggota": "002", "nama_anggota": " siti aminah ", "email": "siti@e2e.test", "password": "rahasia123", "password_confirmation": "rahasia123"})
    call("register.dupemail", "POST", "/api/v1/auth/register", json_body={
        "no_anggota": "0123", "nama_anggota": 'Andi "Si Cepat"', "email": "siti@e2e.test", "password": "rahasia123", "password_confirmation": "rahasia123"})

    # ---------- settings ----------
    call("settings.update.member", "POST", "/api/v1/system/settings/contact", "budi", json_body={"allow_member_registration": True})
    call("settings.update.invalid", "POST", "/api/v1/system/settings/contact", "admin", json_body={"contact_email": "bad", "allow_member_registration": "yes"})
    call("settings.update", "POST", "/api/v1/system/settings/contact", "admin", json_body={
        "contact_email": "halo@e2e.test", "contact_phone": "0812", "contact_address": "Jl. Test", "allow_member_registration": False})
    call("settings.contact.after", "GET", "/api/v1/system/settings/contact")
    call("register.disabled", "POST", "/api/v1/auth/register", json_body={
        "no_anggota": "0123", "nama_anggota": 'Andi "Si Cepat"', "email": "andi@e2e.test", "password": "rahasia123", "password_confirmation": "rahasia123"})

    # ---------- users ----------
    call("users.list.member", "GET", "/api/v1/system/users", "budi")
    call("users.list", "GET", "/api/v1/system/users", "admin")
    call("users.show.self", "GET", f"/api/v1/system/users/{BUDI}", "budi")
    call("users.show.other", "GET", f"/api/v1/profile/{ADMIN}", "budi")
    call("users.show.missing", "GET", f"/api/v1/system/users/{MISSING}", "admin")
    call("users.create.invalid", "POST", "/api/v1/system/users", "admin", json_body={"name": "", "email": "admin@e2e.test", "password": "1", "role": "god"})
    e = call("users.create", "POST", "/api/v1/system/users", "admin", data={
        "name": "Baru", "email": "baru@e2e.test", "password": "rahasia123", "role": "officer", "member_id": "c0000000-0000-4000-8000-000000000003"},
        files={"avatar": ("a.png", PNG, "image/png")})
    new_user = (data_of(e) or {}).get("id")
    call("member.show.linked", "GET", "/api/v1/content/member/c0000000-0000-4000-8000-000000000003")
    call("users.update.self.json", "PUT", f"/api/v1/profile/{BUDI}", "budi", json_body={
        "name": "Budi S", "email": "budi@e2e.test", "kontak": "0899", "domisili": None, "role": "admin"})
    call("users.update.self.form", "POST", f"/api/v1/system/users/{BUDI}", "budi", data={
        "_method": "PUT", "name": "Budi Santoso", "email": "budi@e2e.test", "instagram": "@budi2"},
        files={"avatar": ("b.png", PNG, "image/png")})
    call("users.update.other", "PUT", f"/api/v1/system/users/{ADMIN}", "budi", json_body={"name": "x", "email": "y@z.c"})
    call("users.update.admin.invalid", "PUT", f"/api/v1/system/users/{LAIN}", "admin", json_body={"name": "Lain", "email": "budi@e2e.test"})
    call("users.update.admin", "PUT", f"/api/v1/system/users/{LAIN}", "admin", json_body={"name": "Lain", "email": "lain@e2e.test", "role": "officer"})
    call("users.delete.self", "DELETE", f"/api/v1/system/users/{ADMIN}", "admin")
    call("users.delete.missing", "DELETE", f"/api/v1/system/users/{MISSING}", "admin")
    if new_user:
        call("users.delete", "DELETE", f"/api/v1/system/users/{new_user}", "admin")

    # ---------- articles ----------
    call("article.create.invalid", "POST", "/api/v1/content/article", "budi", json_body={"title": "x", "image": "not-url", "status": "weird"})
    call("article.create.officer", "POST", "/api/v1/content/article", "officer", json_body={"title": "x"})
    call("article.create.url", "POST", "/api/v1/content/article", "budi", json_body={
        "title": "Artikel Baru Budi", "body": "<p>baru</p>", "image": "https://example.com/n.jpg", "published_at": "2025-06-01T10:00:00.000Z"})
    time.sleep(1.1)
    call("article.create.draft", "POST", "/api/v1/content/article", "budi", json_body={"status": "draft"})
    time.sleep(1.1)
    call("article.create.file", "POST", "/api/v1/content/article", "admin", data={"title": "Dengan Gambar", "body": "isi"},
         files={"image": ("pic.png", PNG, "image/png")})
    call("article.list.budi.own", "GET", f"/api/v1/content/article?user_id={BUDI}", "budi")
    call("article.list.budi.other", "GET", f"/api/v1/content/article?user_id={ADMIN}", "budi")
    call("article.list.admin", "GET", "/api/v1/content/article?limit=20", "admin")
    call("article.show.draft.owner", "GET", "/api/v1/content/article/draft-budi-1700000001", "budi")
    call("article.update.missing", "PUT", f"/api/v1/content/article/{MISSING}", "admin", json_body={"title": "x", "body": "y"})
    call("article.update.invalid", "PUT", f"/api/v1/content/article/{ART_BUDI}", "budi", json_body={"title": None})
    call("article.update", "PUT", f"/api/v1/content/article/{ART_BUDI}", "budi", json_body={"title": "Artikel Budi Revisi", "body": "baru", "published_at": "2025-02-02 10:00:00"})
    call("article.update.form", "POST", f"/api/v1/content/article/{ART_DRAFT}", "budi", data={"_method": "PUT", "title": "Draft Budi", "body": "isi", "status": "active"},
         files={"image": ("x.png", PNG, "image/png")})
    call("article.update.todraft", "PUT", f"/api/v1/content/article/{ART1}", "admin", json_body={"status": "draft"})
    call("article.show.id.after", "GET", f"/api/v1/content/article/id/{ART_BUDI}", "admin")
    call("article.show.id.after2", "GET", f"/api/v1/content/article/id/{ART_DRAFT}", "admin")
    call("article.show.id.after3", "GET", f"/api/v1/content/article/id/{ART1}", "admin")
    call("article.delete.missing", "DELETE", f"/api/v1/content/article/{MISSING}", "admin")
    call("article.delete", "DELETE", f"/api/v1/content/article/{ART_FUTURE}", "admin")

    # ---------- news ----------
    call("news.create.member", "POST", "/api/v1/content/news", "budi", json_body={"title": "x"})
    call("news.create.invalid", "POST", "/api/v1/content/news", "officer", json_body={"title": "x"})
    call("news.create", "POST", "/api/v1/content/news", "officer", data={"title": "Berita Tiga", "body": "tiga", "published_at": "2025-04-01"},
         files={"image": ("n.png", PNG, "image/png")})
    call("news.create.badfile", "POST", "/api/v1/content/news", "officer", data={"title": "Berita Empat", "body": "x"},
         files={"image": ("n.png", HTML, "image/png")})
    call("news.update.invalid", "PUT", f"/api/v1/content/news/{NEWS2}", "officer", json_body={"body": "x"})
    call("news.update", "PUT", f"/api/v1/content/news/{NEWS2}", "officer", json_body={"title": "Berita Dua Baru", "body": "x", "published_at": None})
    call("news.show.id.after", "GET", f"/api/v1/content/news/id/{NEWS2}")
    call("news.list.after", "GET", "/api/v1/content/news?filter=asc")
    call("news.delete", "DELETE", f"/api/v1/content/news/{NEWS1}", "admin")

    # ---------- gallery ----------
    call("gallery.create.noauth", "POST", "/api/v1/content/gallery", json_body={})
    call("gallery.create.invalid", "POST", "/api/v1/content/gallery", "budi", data={"caption": "x", "latitude": "200"},
         files={"image": ("g.png", HTML, "image/png")})
    e = call("gallery.create", "POST", "/api/v1/content/gallery", "budi", data={"caption": "Kegiatan Diklat 2026", "latitude": "-6.9", "longitude": "107.6"},
             files={"image": ("g.png", PNG, "image/png")})
    gid = (data_of(e) or {}).get("id")
    if gid:
        call("gallery.update", "POST", f"/api/v1/content/gallery/{gid}", "budi", data={"_method": "PUT", "caption": "Diklat Revisi", "latitude": "-7"})
        call("gallery.show.after", "GET", f"/api/v1/content/gallery/{gid}")
        call("gallery.delete", "DELETE", f"/api/v1/content/gallery/{gid}", "budi")

    # ---------- members ----------
    call("member.create.invalid", "POST", "/api/v1/content/member", "admin", json_body={"no_anggota": "001", "angkatan_id": "nope"})
    e = call("member.create", "POST", "/api/v1/content/member", "admin", json_body={
        "no_anggota": "777", "nama_lengkap": "Anggota Baru", "angkatan_id": ANG2, "kontak": "0812", "avatar": "https://example.com/a.jpg"})
    mid = (data_of(e) or {}).get("id")
    if mid:
        call("member.update", "PUT", f"/api/v1/content/member/{mid}", "admin", json_body={
            "no_anggota": "777", "nama_lengkap": "Anggota Baru 2", "angkatan_id": ANG1, "domisili": "Garut"})
        call("member.update.form", "POST", f"/api/v1/content/member/{mid}", "admin", data={
            "_method": "PUT", "no_anggota": "778", "nama_lengkap": "Anggota Baru 3", "angkatan_id": ANG1},
            files={"avatar": ("m.png", PNG, "image/png")})
        call("member.list.after", "GET", "/api/v1/content/member")
        call("member.delete", "DELETE", f"/api/v1/content/member/{mid}", "admin")
    call("member.delete.missing", "DELETE", f"/api/v1/content/member/{MISSING}", "admin")

    # ---------- angkatan ----------
    call("angkatan.create.dup", "POST", "/api/v1/content/angkatan", "admin", json_body={"nama_angkatan": "Angkatan XXI", "year": "20211"})
    e = call("angkatan.create", "POST", "/api/v1/content/angkatan", "admin", json_body={"nama_angkatan": "Angkatan XXII", "year": "2022"})
    aid = (data_of(e) or {}).get("id")
    if aid:
        call("angkatan.update.dup", "PUT", f"/api/v1/content/angkatan/{aid}", "admin", json_body={"nama_angkatan": "Angkatan XXI", "year": "2022"})
        call("angkatan.update", "PUT", f"/api/v1/content/angkatan/{aid}", "admin", json_body={"nama_angkatan": "Angkatan XXII", "year": "2023"})
        call("angkatan.delete", "DELETE", f"/api/v1/content/angkatan/{aid}", "admin")

    # ---------- comments ----------
    call("comment.create.noauth", "POST", "/api/v1/content/comments", json_body={"body": "x"})
    call("comment.create.invalid", "POST", "/api/v1/content/comments", "budi", json_body={"article_id": MISSING})
    e = call("comment.create", "POST", "/api/v1/content/comments", "budi", json_body={"body": "Komentar baru", "news_id": NEWS2})
    cid = (data_of(e) or {}).get("id")
    time.sleep(1.1)
    call("comment.reply", "POST", "/api/v1/content/comments", "officer", json_body={"body": "Balas", "news_id": NEWS2, "parent_id": cid})
    call("comments.list.news", "GET", f"/api/v1/content/comments?news_id={NEWS2}")
    call("comment.delete.other", "DELETE", f"/api/v1/content/comments/{cid}", "lain")
    call("comment.delete", "DELETE", f"/api/v1/content/comments/{cid}", "budi")
    call("comment.delete.admin", "DELETE", f"/api/v1/content/comments/{C1}", "admin")
    call("comments.list.after", "GET", f"/api/v1/content/comments?article_id={ART1}")

    # ---------- logout ----------
    call("logout", "POST", "/api/v1/auth/logout", "budi")
    call("auth.me.afterlogout", "GET", "/api/v1/auth/current-user", "budi")
    return results


UUID = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
SEEDED = re.compile(r"^(a0|b0|c0|d0|e0|f0|99|70|00)000000-")
TODAY = time.strftime("%Y-%m-%d", time.gmtime())


def normalize(obj, ids):
    if isinstance(obj, dict):
        return {k: normalize(v, ids) for k, v in obj.items()}
    if isinstance(obj, list):
        return [normalize(v, ids) for v in obj]
    if isinstance(obj, str):
        s = obj
        s = re.sub(r"http://127\.0\.0\.1:81\d\d", "<host>", s)
        s = re.sub(r"\b\d+\|[A-Za-z0-9]{48}\b", "<token>", s)
        s = re.sub(r"[A-Za-z0-9]{30}\.(png|jpg|jpeg|gif|svg)", r"<file>.\1", s)
        s = re.sub(r"-1\d{9}\b", "-<ts>", s)
        if TODAY in s or re.match(r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}", s) and TODAY in s:
            s = "<now>"

        def rep(m):
            u = m.group(0)
            if SEEDED.match(u):
                return u
            ids.setdefault(u, f"<uuid{len(ids) + 1}>")
            return ids[u]
        return UUID.sub(rep, s)
    return obj


def diff(a, b, path="", out=None):
    out = [] if out is None else out
    if isinstance(a, dict) and isinstance(b, dict):
        for k in a.keys() - b.keys():
            out.append(f"{path}.{k}: only in Laravel = {json.dumps(a[k])[:120]}")
        for k in b.keys() - a.keys():
            out.append(f"{path}.{k}: only in Go (extra) = {json.dumps(b[k])[:120]}")
        for k in a.keys() & b.keys():
            diff(a[k], b[k], f"{path}.{k}", out)
    elif isinstance(a, list) and isinstance(b, list):
        if len(a) != len(b):
            out.append(f"{path}: list length {len(a)} vs {len(b)}")
        for i, (x, y) in enumerate(zip(a, b)):
            diff(x, y, f"{path}[{i}]", out)
    elif a != b:
        out.append(f"{path}: Laravel={json.dumps(a)[:150]} Go={json.dumps(b)[:150]}")
    return out


if __name__ == "__main__":
    lar, go = scenario(sys.argv[1]), scenario(sys.argv[2])
    assert any("E2E-ANGKATAN" in json.dumps(e["body"]) for n, e in lar if n == "angkatan.list"), "Laravel not on e2e DB!"
    ids_l, ids_g = {}, {}
    n_same = 0
    report = []
    for (name, l), (_, g) in zip(lar, go):
        issues = []
        if l["status"] != g["status"]:
            issues.append(f"status Laravel={l['status']} Go={g['status']}")
        if isinstance(l["body"], bytes):
            if l["ctype"] != g["ctype"]:
                issues.append(f"content-type {l['ctype']} vs {g['ctype']}")
            if name in ("export.csv", "image.show") and l["body"] != g["body"]:
                issues.append(f"bytes differ:\n  L={l['body'][:200]!r}\n  G={g['body'][:200]!r}")
            if name.startswith("share"):
                lb = re.sub(r"\s+", " ", l["body"].decode(errors="replace")).strip()
                gb = re.sub(r"\s+", " ", g["body"].decode(errors="replace")).strip()
                if lb != gb:
                    issues.append(f"html differs:\n  L={lb[:400]}\n  G={gb[:400]}")
        else:
            issues += diff(normalize(l["body"], ids_l), normalize(g["body"], ids_g), "")
        if issues:
            report.append(f"### {name}\n  " + "\n  ".join(issues))
        else:
            n_same += 1
    print(f"{n_same}/{len(lar)} identical\n")
    print("\n".join(report))
    json.dump({"laravel": [(n, e if not isinstance(e["body"], bytes) else {**e, "body": repr(e["body"][:300])}) for n, e in lar],
               "go": [(n, e if not isinstance(e["body"], bytes) else {**e, "body": repr(e["body"][:300])}) for n, e in go]},
              open(sys.argv[3], "w"), indent=1)
