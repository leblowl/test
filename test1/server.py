import json
from http.server import BaseHTTPRequestHandler, HTTPServer

class WebRequestHandler(BaseHTTPRequestHandler):

    data_len = 0
    data = []
    # val1: [dataNdx0, dataNdx1]
    # val2: [dataNdx2, dataNdx3]
    data_index = {}
    # val1: part1
    # val2: part1
    part_index = {}
    part = file_next()

    def addElem[e]:
        data.append(e)

        len(index) > 1000:
          # write data_index to part
          data_index = {}
          part = file_next()

        # see if there is a data_index existing
        if e in part_index:
            data_index = get_partition_from_disk(part_index[e])
            data_index[e].append(data_len - 1)
            data_len += 1



            # get data position
            data_len
            if part_index[e]
        if e in index:
            data_index[e].append(data.length - 1)
        else:
            data_index[e] = [data.length - 1]

        # write index to file

    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps([1, 2, 3]).encode("utf-8"))


if __name__ == "__main__":
    server = HTTPServer(("0.0.0.0", 8000), WebRequestHandler)
    server.serve_forever()


    # "abc": [0, 2, 4]
    # "abcd": [5]
    # ...
    # "abcde": [7, 8]
