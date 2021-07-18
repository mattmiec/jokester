import React from 'react';
import $ from 'jquery';

export class LoggedIn extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      jokes: []
    };

    this.serverRequest = this.serverRequest.bind(this);
    this.logout = this.logout.bind(this);
    this.likeJoke = this.likeJoke.bind(this);
    this.unlikeJoke = this.unlikeJoke.bind(this);
  }

  logout() {
    localStorage.removeItem("id_token");
    localStorage.removeItem("access_token");
    localStorage.removeItem("profile");
    location.reload();
  }

  serverRequest() {
    $.get("http://localhost:3000/api/jokes", res => {
      this.setState({
        jokes: res
      });
    });
  }

  componentDidMount() {
    this.serverRequest();
  }

  likeJoke(i) {
    $.post(
      "http://localhost:3000/api/jokes/like/" + this.state.jokes[i].id,
      {},
      res => {
        console.log("res... ", res);
        const newjokes = this.state.jokes.map(j => Object.assign({}, j));
        newjokes[i].liked = true;
        this.setState({ jokes: newjokes });
      });
  }

  unlikeJoke(i) {
    $.post(
      "http://localhost:3000/api/jokes/unlike/" + this.state.jokes[i].id,
      {},
      res => {
        console.log("res... ", res);
        const newjokes = this.state.jokes.map(j => Object.assign({}, j));
        newjokes[i].liked = false;
        this.setState({ jokes: newjokes });
      });
  }

  render() {
    return (
      <div className="container">
        <br />
        <span className="float-right">
          <button className onClick={this.logout} className="btn btn-secondary">Log out</button>
        </span>
        <h2>Jokester</h2>
        <p>Here are some jokes!</p>
        <div className="row align-items-start">
          {this.state.jokes.map((joke, i) => {
            return <Joke key={`${joke.id}-${joke.liked}`} joke={joke} action={() => joke.liked ? this.unlikeJoke(i) : this.likeJoke(i)}/>;
          })}
        </div>
      </div>
    );
  }
}

function Joke(props) {
    const date = new Date(props.joke.created);
    return (
      <div className="col-xs-4">
        <div className="card">
          <div className="card-header">
            {`by ${props.joke.author} at ${date.toLocaleTimeString()} on ${date.toLocaleDateString()}`}
          </div>
          <div className="card-body">
              {props.joke.joke}
          </div>
          <div className="card-footer">
            {props.joke.likes} Likes &nbsp;
            <button onClick={props.action} className="btn btn-outline-secondary">
                <i className={props.joke.liked ? "fas fa-check" : "fas fa-thumbs-up"} aria-hidden="true"></i>
            </button>
          </div>
        </div>
      </div>
    )
}