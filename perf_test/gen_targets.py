import argparse
import base64
import json
import os
import random
import string

AUTHOR_IDS = list(range(900_000, 900_050))

TITLE_ALPHABET = string.ascii_letters + string.digits + " "
TEXT_ALPHABET = string.ascii_letters + string.digits + " .,!?-"


def random_text(rnd: random.Random, alphabet: str, min_len: int, max_len: int) -> str:
    length = rnd.randint(min_len, max_len)
    return "".join(rnd.choice(alphabet) for _ in range(length)).strip() or "x"


def gen_create_targets(path: str, host: str, count: int, seed: int) -> None:
    rnd = random.Random(seed)
    url = f"{host}/v1/posts"
    header = {"Content-Type": ["application/json"]}
    with open(path, "w", encoding="utf-8") as f:
        for _ in range(count):
            body = {
                "author_user_id": rnd.choice(AUTHOR_IDS),
                "title": random_text(rnd, TITLE_ALPHABET, 10, 60),
                "blocks": [
                    {
                        "kind": "CONTENT_BLOCK_KIND_TEXT",
                        "text_content": random_text(rnd, TEXT_ALPHABET, 40, 400),
                    }
                ],
            }
            raw = json.dumps(body, ensure_ascii=False).encode("utf-8")
            target = {
                "method": "POST",
                "url": url,
                "header": header,
                "body": base64.b64encode(raw).decode("ascii"),
            }
            f.write(json.dumps(target, ensure_ascii=False) + "\n")


def gen_read_targets(path: str, host: str, count: int, max_id: int, seed: int) -> None:
    rnd = random.Random(seed + 1)
    with open(path, "w", encoding="utf-8") as f:
        for _ in range(count):
            post_id = rnd.randint(1, max_id)
            f.write(f"GET {host}/v1/posts/{post_id}\n")


def main() -> None:
    parser = argparse.ArgumentParser(description="Генератор vegeta-таргетов для постов")
    parser.add_argument("--count", type=int, default=100_000)
    parser.add_argument("--read-count", type=int, default=10_000)
    parser.add_argument("--max-id", type=int, default=100_000)
    parser.add_argument("--host", default="http://localhost:8083")
    parser.add_argument("--out-dir", default=".")
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    create_path = os.path.join(args.out_dir, "create_targets.json")
    read_path = os.path.join(args.out_dir, "read_targets.txt")

    gen_create_targets(create_path, args.host, args.count, args.seed)
    gen_read_targets(read_path, args.host, args.read_count, args.max_id, args.seed)

    print(f"created {args.count} POST targets  -> {create_path}")
    print(f"created {args.read_count} GET targets   -> {read_path}")


if __name__ == "__main__":
    main()
