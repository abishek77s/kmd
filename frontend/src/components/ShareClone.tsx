import { useState } from "react";

const ShareClone = () => {
  const [name, setName] = useState("");
  const [content, setContent] = useState("");
  const [id, setId] = useState("");
  const [inputId, setInputId] = useState("");

  const [selected, setSelected] = useState(0);

  interface File {
    id: string;
    name: string;
    content: string;
  }

  const [data, setData] = useState<File>();
  const [loading, setLoading] = useState(false);

  const handleGet = async () => {
    setLoading(true);
    try {
      const res = await fetch(
        "https://kmd-backend.onrender.com/file/" + inputId,
      );
      const json = await res.json();
      setData(json);
    } catch (err) {
      console.error("Error fetching:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    console.log(id, name, content);
    try {
      const id = name + "-" + Math.floor(Math.random() * 1000).toString();
      console.log(id);

      const res = await fetch("https://kmd-backend.onrender.com/file", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          id,
          name,
          content,
        }),
      });

      const data = await res.json();
      setId(id.toString());
      try {
        await navigator.clipboard.writeText(id);
        // optional: toast/snackbar
        console.log("Copied!");
      } catch (err) {
        console.error("Copy failed:", err);
      }
      console.log("Success:", data);
    } catch (error) {
      console.error("Error:", error);
    }
  };

  return (
    <>
      <div className="flex space-x-2 pb-4 ">
        <button
          onClick={() => setSelected(0)}
          className={`transition-all duration-200  font-mono rounded-sm py-2 text-center
              ${selected === 0 ? "px-8 bg-blue-400 text-white font-bold" : "px-3 bg-white"}`}
        >
          Share
        </button>
        <button
          onClick={() => setSelected(1)}
          className={`transition-all duration-200 font-mono rounded-sm py-2
            ${selected === 1 ? "px-8 bg-pink-400 text-white font-bold" : "px-3 bg-white"}`}
        >
          Clone
        </button>
      </div>
      <div className="">
        {selected == 0 && (
          <div className="flex flex-col bg-zinc-100 h-96 w-11/12">
            <input
              className="bg-zinc-300"
              type="text"
              placeholder="File name"
              onChange={(e) => setName(e.target.value)}
            />
            <input
              className="h-full"
              onChange={(e) => setContent(e.target.value)}
              type="text"
              placeholder="Content"
            />
            {id && <h1> {id}</h1>}
            <button
              onClick={handleSubmit}
              className={`transition-all duration-200 font-mono rounded-sm py-2 px-8
                        bg-blue-400 text-white font-bold text-center
                        active:scale-105`}
            >
              Get ID
            </button>
          </div>
        )}
        {selected == 1 && (
          <div className="flex flex-col bg-zinc-100  w-11/12">
            <input
              className="bg-zinc-300"
              onChange={(e) => setInputId(e.target.value)}
              type="text"
              placeholder="Paste ID"
            />

            <button
              onClick={handleGet}
              className={`transition-all duration-200 font-mono rounded-sm py-2 px-8
                      bg-blue-400 text-white font-bold text-center
                      active:scale-105`}
            >
              Get
            </button>
            {loading && <p> Loading...</p>}
            <h1>{data?.name}</h1>
            <p>{data?.content}</p>
          </div>
        )}
      </div>
    </>
  );
};

export default ShareClone;
