import axios from 'axios';
import React, { Component } from 'react';
import Badge from 'react-bootstrap/Badge';
import Button from 'react-bootstrap/Button';
import Col from 'react-bootstrap/Col';
import Row from 'react-bootstrap/Row';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Spinner from 'react-bootstrap/Spinner';
import { handleChange } from './forms';
import { SetupRepository } from './SetupRepository';
import { CLIEquivalent } from './uiutil';

export class RepoStatus extends Component {
    constructor() {
        super();

        this.state = {
            status: {},
            isLoading: true,
            error: null,
            provider: "",
            description: "",
        };

        this.mounted = false;
        this.disconnect = this.disconnect.bind(this);
        this.updateDescription = this.updateDescription.bind(this);
        this.handleChange = handleChange.bind(this);
    }

    componentDidMount() {
        this.mounted = true;
        this.fetchStatus(this.props);
    }

    componentWillUnmount() {
        this.mounted = false;
    }

    fetchStatus(props) {
        if (this.mounted) {
            this.setState({
                isLoading: true,
            });
        }

        axios.get('/api/v1/repo/status').then(result => {
            if (this.mounted) {
                this.setState({
                    status: result.data,
                    isLoading: false,
                });
            }
        }).catch(error => {
            if (this.mounted) {
                this.setState({
                    error,
                    isLoading: false
                })
            }
        });
    }

    disconnect() {
        this.setState({ isLoading: true })
        axios.post('/api/v1/repo/disconnect', {}).then(result => {
            window.location.replace("/");
        }).catch(error => this.setState({
            error,
            isLoading: false
        }));
    }

    selectProvider(provider) {
        this.setState({ provider });
    }

    updateDescription() {
        this.setState({
            isLoading: true
        });

        axios.post('/api/v1/repo/description', {
            "description": this.state.status.description,
        }).then(result => {
            this.setState({
                isLoading: false,
            });
        }).catch(error => {
            this.setState({
                isLoading: false,
            });
        });
    }

    render() {
        let { isLoading, error } = this.state;
        if (error) {
            return <p>ERROR: {error.message}</p>;
        }
        if (isLoading) {
            return <Spinner animation="border" variant="primary" />;
        }

        return this.state.status.connected ?
            <>
                <h3>Connected To Repository</h3>
                <Form onSubmit={this.updateDescription}>
                    <Row>
                        <Form.Group as={Col}>
                            <InputGroup>
                                <Form.Control
                                    autoFocus="true"
                                    isInvalid={!this.state.status.description}
                                    name="status.description"
                                    value={this.state.status.description}
                                    onChange={this.handleChange}
                                    size="sm" />
                                &nbsp;
                                <Button size="sm" type="submit">Update Description</Button>
                            </InputGroup>
                            <Form.Control.Feedback type="invalid">Description Is Required</Form.Control.Feedback>
                        </Form.Group>
                    </Row>
                    {this.state.status.readonly && <Row>
                        <Badge pill variant="warning">Repository is read-only</Badge>
                    </Row>}
                </Form>
                <hr />
                <Form>
                    {this.state.status.apiServerURL ? <>
                        <Row>
                            <Form.Group as={Col}>
                                <Form.Label>Server URL</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.apiServerURL} />
                            </Form.Group>
                        </Row>
                    </> : <>
                        <Row>
                            <Form.Group as={Col}>
                                <Form.Label>Config File</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.configFile} />
                            </Form.Group>
                        </Row>
                        <Row>
                            <Form.Group as={Col}>
                                <Form.Label>Provider</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.storage} />
                            </Form.Group>
                            <Form.Group as={Col}>
                                <Form.Label>Hash Algorithm</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.hash} />
                            </Form.Group>
                            <Form.Group as={Col}>
                                <Form.Label>Encryption Algorithm</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.encryption} />
                            </Form.Group>
                            <Form.Group as={Col}>
                                <Form.Label>Splitter Algorithm</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.splitter} />
                            </Form.Group>
                            <Form.Group as={Col}>
                                <Form.Label>Internal Compression</Form.Label>
                                <Form.Control readOnly defaultValue={this.state.status.supportsContentCompression ? "yes" : "no"} />
                            </Form.Group>
                        </Row>
                    </>}
                    <Row>
                        <Form.Group as={Col}>
                            <Form.Label>Connected as:</Form.Label>
                            <Form.Control readOnly defaultValue={this.state.status.username + "@" + this.state.status.hostname} />
                        </Form.Group>
                    </Row>
                    <Row><Col>&nbsp;</Col></Row>
                    <Row>
                        <Col>
                            <Button size="sm" variant="danger" onClick={this.disconnect}>Disconnect</Button>
                        </Col>
                    </Row>
                </Form>
                <Row><Col>&nbsp;</Col></Row>
                <Row>
                    <Col xs={12}>
                        <CLIEquivalent command="repository status" />
                    </Col>
                </Row>
            </> : <SetupRepository />
    }
}
